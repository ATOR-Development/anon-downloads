package downloads

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/ATOR-Development/anon-download-links/internal/config"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type Artifact struct {
	Name        string `json:"name"`
	DownloadURL string `json:"download_url"`
}

type Downloads struct {
	owner       string
	repo        string
	token       string
	releases    []*release
	cachePeriod time.Duration

	logger log.Logger

	latestUpdate         time.Time
	latestArtifacts      []*Artifact
	latestArtifactsMutex sync.Mutex
}

func New(cfg *config.Config, logger log.Logger) (*Downloads, error) {
	var releases []*release
	for _, r := range cfg.Artifacts {
		regexp, err := regexp.Compile(r.Regexp)
		if err != nil {
			return nil, fmt.Errorf("release regexp (%s): %w", r.Name, err)
		}

		releases = append(releases, &release{
			name:   r.Name,
			regexp: regexp,
		})
	}

	cachePeriod, err := time.ParseDuration(cfg.CachePeriod)
	if err != nil {
		return nil, fmt.Errorf("cache period parse: %w", err)
	}

	return &Downloads{
		owner:       cfg.Owner,
		repo:        cfg.Repo,
		token:       cfg.Token,
		releases:    releases,
		cachePeriod: cachePeriod,

		logger: logger,
	}, nil
}

func (d *Downloads) GetArtifacts(ctx context.Context) ([]*Artifact, error) {
	// TODO: Cache results for N seconds/minutes
	return d.fetchArtifacts(ctx)
}

func (d *Downloads) GetArtifactsMap(ctx context.Context) (map[string]string, error) {
	artifacts, err := d.fetchArtifacts(ctx)
	if err != nil {
		return nil, err
	}

	artifactsMap := make(map[string]string)
	for _, a := range artifacts {
		artifactsMap[a.Name] = a.DownloadURL
	}

	return artifactsMap, nil
}

func (d *Downloads) fetchArtifacts(ctx context.Context) ([]*Artifact, error) {
	d.latestArtifactsMutex.Lock()
	defer d.latestArtifactsMutex.Unlock()

	if d.latestUpdate.Add(d.cachePeriod).Compare(time.Now()) > 0 {
		d.logger.Log("msg", "cache hit")
		return d.latestArtifacts, nil
	} else {
		d.logger.Log("msg", "cache miss")
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", d.owner, d.repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", d.token))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var githubReleaseResp githubRelease
	err = json.Unmarshal(respData, &githubReleaseResp)
	if err != nil {
		return nil, err
	}

	var artifacts []*Artifact
	for _, r := range d.releases {
		matches := 0
		var artifact *Artifact
		for _, a := range githubReleaseResp.Assets {
			if r.regexp.MatchString(a.Name) {
				artifact = &Artifact{
					Name:        r.name,
					DownloadURL: a.BrowserDownloadURL,
				}
				matches++
			}
		}

		if artifact != nil {
			if matches != 1 {
				level.Warn(d.logger).Log("msg", "unexpected artifacts count", "name", r.name, "count", matches)
			}
			artifacts = append(artifacts, artifact)
		} else {
			level.Warn(d.logger).Log("msg", "no artifacts found", "name", r.name)
		}
	}

	if len(artifacts) == 0 {
		return nil, errors.New("no matched artifacts were found in latest release")
	}

	d.latestUpdate = time.Now()
	d.latestArtifacts = artifacts

	return artifacts, nil
}

type release struct {
	name   string
	regexp *regexp.Regexp
}
