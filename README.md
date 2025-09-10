# Golancer

**GoLancer** is a lightweight, high-performance reverse proxy and load balancer written in Go.
It is designed to be simple, configurable, and production-ready with clear separation of control plane and data plane, hot-reloadable configuration, structured logging, and graceful shutdown support.

### Features (Completed so far)

- Reverse Proxy
- - Routes HTTP/HTTPS traffic to upstream services.
- - Path prefix and host-based routing.
- Load Balancing
- - Round-robin load balancing across multiple upstreams.
- Dynamic Config Management
- - Configuration defined via YAML.
- - Hot reload powered by Viper + fsnotify.
- - Immutable vs mutable config separation.
- Structured Logging
- - Append-only, file-based logger.
- - Async logging with mailbox pattern (non-blocking).
- Graceful Shutdown
- - Control plane manages shutdown signals.
- - Cleans up data plane, logger, and config watchers.
- TLS/HTTPS Support
- - Self-signed certs for local development.
- - Let’s Encrypt integration for production (planned).

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
