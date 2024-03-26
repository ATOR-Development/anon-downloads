package api

import (
	"encoding/json"
	"net/http"

	"github.com/ATOR-Development/anon-download-links/internal/config"
	"github.com/ATOR-Development/anon-download-links/internal/downloads"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
)

type API struct {
	cfg       *config.Config
	downloads *downloads.Downloads

	logger log.Logger
}

func New(cfg *config.Config, downloads *downloads.Downloads, logger log.Logger) *API {
	return &API{
		cfg:       cfg,
		downloads: downloads,

		logger: log.WithPrefix(logger, "service", "api"),
	}
}

func (a *API) Listen(listenAddress string) {
	router := mux.NewRouter()
	router.HandleFunc("/api/downloads", a.HandleDownloads).Methods("GET")
	router.HandleFunc("/download/{name}", a.HandleDownload).Methods("GET")
	router.HandleFunc("/hc", a.HandleHC).Methods("GET")

	http.Handle("/", router)
	http.ListenAndServe(listenAddress, nil)
}

func (a *API) HandleDownloads(w http.ResponseWriter, r *http.Request) {
	level.Error(a.logger).Log("msg", "handling downloads")

	artifacts, err := a.downloads.GetArtifacts(r.Context())
	if err != nil {
		level.Error(a.logger).Log("msg", "unable to get artifacts", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respBytes, err := json.Marshal(artifacts)
	if err != nil {
		level.Error(a.logger).Log("msg", "unable to marshal artifacts", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(respBytes)
}

func (a *API) HandleDownload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	level.Error(a.logger).Log("msg", "handling download", "name", name)

	artifactsMap, err := a.downloads.GetArtifactsMap(r.Context())
	if err != nil {
		level.Error(a.logger).Log("msg", "unable to get artifacts map", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	downloadURL, ok := artifactsMap[name]
	if !ok {
		level.Warn(a.logger).Log("msg", "download not found", "name", name)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("download not found"))
		return
	}

	http.Redirect(w, r, downloadURL, http.StatusFound)
}

func (a *API) HandleHC(w http.ResponseWriter, r *http.Request) {
	level.Error(a.logger).Log("msg", "handling health check")

	artifacts, err := a.downloads.GetArtifacts(r.Context())
	if err != nil {
		level.Error(a.logger).Log("msg", "unable to get artifacts", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(artifacts) != len(a.cfg.Artifacts) {
		level.Error(a.logger).Log("msg", "unable to fetch all artifacts", "required", len(a.cfg.Artifacts), "fetched", len(artifacts))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte("OK"))
}
