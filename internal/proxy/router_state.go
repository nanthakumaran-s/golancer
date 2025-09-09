package proxy

import (
	"net/http"
	"time"
)

type Route struct {
	Name       string
	Hosts      []string
	PathPrefix string
	Pool       *UpstreamPool
}

type RouterState struct {
	Routes    []Route
	Transport *http.Transport
	Timeout   time.Duration
}
