package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"

	"github.com/ATOR-Development/anon-download-links/internal/api"
	"github.com/ATOR-Development/anon-download-links/internal/config"
	"github.com/ATOR-Development/anon-download-links/internal/downloads"
)

var (
	configFile    = flag.String("config", "config.yml", "Config file.")
	listenAddress = flag.String("listen-address", ":8080", "Exporter HTTP listen address.")
)

func main() {
	flag.Parse()

	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.WithPrefix(logger, "ts", log.TimestampFormat(time.Now, time.Stamp))

	level.Info(logger).Log("msg", "initializing service from", "config", *configFile)

	cfg, err := config.FromFile(*configFile)
	if err != nil {
		level.Error(logger).Log("msg", "cannot read config", "err", err.Error())
		os.Exit(1)
	}

	downloads, err := downloads.New(cfg, logger)
	if err != nil {
		level.Error(logger).Log("msg", "cannot create downloads service", "err", err.Error())
		os.Exit(1)
	}

	level.Info(logger).Log("msg", "starting http server", "listen", *listenAddress)

	api := api.New(downloads, logger)

	router := mux.NewRouter()
	router.HandleFunc("/api/downloads", api.HandleDownloads).Methods("GET")
	router.HandleFunc("/download/{name}", api.HandleDownload).Methods("GET")

	http.Handle("/", router)
	http.ListenAndServe(*listenAddress, nil)
}
