package downloads

import "time"

type githubRelease struct {
	HTMLURL     string               `json:"html_url"`
	TagName     string               `json:"tag_name"`
	CreatedAt   time.Time            `json:"created_at"`
	PublishedAt time.Time            `json:"published_at"`
	Assets      []githubReleaseAsset `json:"assets"`
}

type githubReleaseAsset struct {
	Name               string `json:"name"`
	Size               int    `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
}
