# Anon Downloads

This service provides API and permanent links to download anon packages from GitHub releases.

### Available command line options:

```
Usage of ./bin/service:
  -config string
    	Config file. (default "config.yml")
  -listen-address string
    	Exporter HTTP listen address. (default ":8080")
```

## Build

Make sure you have Go installed and it is in your `PATH`.

```
make build
```

## Run

Make sure you created `config.yml` before running.

```
make run
```