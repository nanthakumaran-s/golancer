package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type Router struct {
	State *RouterState
}

func NewTransport(maxIdle int, idleTimeout time.Duration) *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          maxIdle,
		IdleConnTimeout:       idleTimeout,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	st := r.State
	rt := match(st.Routes, req)
	if rt == nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	be := rt.Pool.Next()
	if be == nil {
		http.Error(w, "No upstream", http.StatusBadGateway)
		return
	}

	out := req.Clone(req.Context())
	out.RequestURI = ""
	out.URL.Scheme = be.URL.Scheme
	out.URL.Host = be.URL.Host
	out.URL.Path = join(be.URL.Path, req.URL.Path)

	out.Host = be.URL.Host

	if be.URL.RawQuery == "" || req.URL.RawQuery == "" {
		out.URL.RawQuery = be.URL.RawQuery + req.URL.RawQuery
	} else {
		out.URL.RawQuery = be.URL.RawQuery + "&" + req.URL.RawQuery
	}

	sanitize(out.Header)

	ctx := out.Context()
	if st.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, st.Timeout)
		defer cancel()
		out = out.WithContext(ctx)
	}

	fmt.Printf("Incoming:  %s %s Host=%s\n", req.Method, req.URL.String(), req.Host)
	fmt.Printf("Outgoing:  %s %s Host=%s\n", out.Method, out.URL.String(), out.Host)

	resp, err := st.Transport.RoundTrip(out)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func match(routes []Route, req *http.Request) *Route {
	host := req.Host
	if host == "" && req.URL != nil {
		host = req.URL.Host
	}

	for i := range routes {
		rt := &routes[i]
		if len(rt.Hosts) > 0 {
			ok := false
			for _, h := range rt.Hosts {
				if strings.EqualFold(h, host) {
					ok = true
					break
				}
			}
			if !ok {
				continue
			}
		}
		if rt.PathPrefix != "" && !strings.HasPrefix(req.URL.Path, rt.PathPrefix) {
			continue
		}
		return rt
	}
	return nil
}

var hopByHop = []string{
	"Connection", "Proxy-Connection", "Keep-Alive",
	"Proxy-Authenticate", "Proxy-Authorization",
	"TE", "Trailers", "Transfer-Encoding", "Upgrade",
}

func sanitize(h http.Header) {
	for _, k := range hopByHop {
		h.Del(k)
	}
	if c := h.Get("Connection"); c != "" {
		for _, f := range strings.Split(c, ",") {
			if f = strings.TrimSpace(f); f != "" {
				h.Del(f)
			}
		}
	}
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func join(a, b string) string {
	if a == "" {
		return b
	}
	as := strings.HasSuffix(a, "/")
	bs := strings.HasPrefix(b, "/")
	switch {
	case as && bs:
		return a + b[1:]
	case !as && !bs:
		return a + "/" + b
	default:
		return a + b
	}
}
