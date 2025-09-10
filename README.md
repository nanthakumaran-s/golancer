# Golancer

**GoLancer** is a lightweight, high-performance reverse proxy and load balancer written in Go.
It is designed to be simple, configurable, and production-ready with clear separation of control plane and data plane, hot-reloadable configuration, structured logging, and graceful shutdown support.

### Features (Completed so far)

- Reverse Proxy

  - Routes HTTP/HTTPS traffic to upstream services.
  - Path prefix and host-based routing.

- Load Balancing

  - Round-robin load balancing across multiple upstreams.

- Dynamic Config Management

  - Configuration defined via YAML.
  - Hot reload powered by Viper + fsnotify.
  - Immutable vs mutable config separation.

- Structured Logging

  - Append-only, file-based logger.
  - Async logging with mailbox pattern (non-blocking).

- Graceful Shutdown

  - Control plane manages shutdown signals.
  - Cleans up data plane, logger, and config watchers.

- TLS/HTTPS Support
  - Self-signed certs for local development.
  - Let’s Encrypt integration for production (planned).

## Installation Guide

Clone this repo and run

```
go run ./cmd/golancer start --port 8080 --config config.yaml
```

Flags:

- `--port` or `--p`: Port to run Golancer. `Default: 8080`
- `--config` or `--c`: Config file path to listen. `Default: config.yaml`
- `--useTLS`: Use Golancer in TLS mode. `Default: false`
- `--local`: Set Golancer in local development mode. `Default: false`
- `--logFile` or `--l`: Log file path to write. Default: `golancer.log`

## Configuration Guide

A sample config file looks similar to this

```
proxy:
  default_timeout: 5s
  max_idle_conns: 100
  idle_conn_timeout: 90s

routes:
  - name: api
    match:
      hosts: ["localhost:8080"]
      path_prefix: /api
    upstreams:
      - https://127.0.0.1:8081
      - https://127.0.0.1:8082
    lb: round_robin
```

### Roadmap

- [x] Rever Proxy (HTTP/HTTPS)
- [x] Round Robin Load balancing
- [x] Control plane and Data plane
- [x] Hot reload on configuration changes
- [x] Append-only file based logger
- [ ] More Load balancing strategies (least connection, hash)
- [ ] Metrics (Prometheus exporter)
- [ ] Middleware support
- [ ] Let's encrypt integration for automatic TLS
- [ ] Health checks for upstreams
- [ ] Circuit breakers + retries
- [ ] Graceful reloads without dropping in-flight requests

### Contributing

Contributions, issues, and feature requests are welcome!\
Please open a PR or file an issue.

### License

MIT License © 2025 [Nanthakumaran Senthilnathan]
