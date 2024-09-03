job "anon-downloads-live" {
  datacenters = ["ator-fin"]
  type = "service"
  namespace = "ator-network"

  group "anon-downloads-group" {
    count = 1

    constraint {
      attribute = "${node.unique.id}"
      value     = "c8e55509-a756-0aa7-563b-9665aa4915ab"
    }

    network  {
      mode = "bridge"
      port "downloads-http" {
        static = 8008
        to = 8080
        host_network = "wireguard"
      }
    }

    task "anon-downloads-task" {
      driver = "docker"

      config {
        image = "ghcr.io/ator-development/anon-downloads:v0.0.8"
        ports = ["downloads-http"]
        volumes = [
          "local/config.yml:/app/config.yml:ro",
        ]
      }

      vault {
        policies = ["anon-downloads"]
      }

      resources {
        cpu = 256
        memory = 256
      }

      service {
        name = "anon-downloads"
        port = "downloads-http"
        tags = [
          "traefik.enable=true",
          "traefik.http.routers.anon-downloads.entrypoints=https",
          "traefik.http.routers.anon-downloads.rule=Host(`download-live.dmz.ator.dev`)",
          "traefik.http.routers.anon-downloads.tls=true",
          "traefik.http.routers.anon-downloads.tls.certresolver=atorresolver",

          "traefik.http.routers.any1-downloads.entrypoints=https",
          "traefik.http.routers.any1-downloads.rule=Host(`download.en.anyone.tech`)",
          "traefik.http.routers.any1-downloads.tls=true",
          "traefik.http.routers.any1-downloads.tls.certresolver=anyoneresolver",
        ]
        check {
          name     = "anon downloads alive"
          type     = "http"
          port     = "downloads-http"
          path     = "/hc"
          interval = "10s"
          timeout  = "10s"
          check_restart {
            limit = 10
            grace = "30s"
          }
        }
      }

      template {
        data = <<EOH
owner: ATOR-Development
repo: ator-protocol
{{with secret "kv/anon-downloads"}}
token: "{{.Data.data.GITHUB_TOKEN}}"
{{end}}
cachePeriod: 15m
artifacts:
  - name: macos-amd64
    regexp: '^anon-live-macos-amd64.+'
  - name: macos-arm64
    regexp: '^anon-live-macos-arm64.+'
  - name: windows-amd64
    regexp: '^anon-live-windows-amd64.+'
  - name: debian-bullseye-amd64
    regexp: '^anon.+-live-.+bullseye.+amd64\.deb'
  - name: debian-bullseye-arm64
    regexp: '^anon.+-live-.+bullseye.+arm64\.deb'
  - name: debian-bookworm-amd64
    regexp: '^anon.+-live-.+bookworm.+amd64\.deb'
  - name: debian-bookworm-arm64
    regexp: '^anon.+-live-.+bookworm.+arm64\.deb'
  - name: ubuntu-focal-amd64
    regexp: '^anon.+-live-.+focal.+amd64\.deb'
  - name: ubuntu-focal-arm64
    regexp: '^anon.+-live-.+focal.+arm64\.deb'
  - name: ubuntu-jammy-amd64
    regexp: '^anon.+-live-.+jammy.+amd64\.deb'
  - name: ubuntu-jammy-arm64
    regexp: '^anon.+-live-.+jammy.+arm64\.deb'
        EOH
        destination = "local/config.yml"
      }
    }
  }
}
