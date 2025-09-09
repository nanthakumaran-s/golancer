package proxy

import (
	"net/http"
	"sync/atomic"
)

type AtomicHandler struct {
	v atomic.Value
}

func NewAtomicHandler(h http.Handler) *AtomicHandler {
	a := &AtomicHandler{}
	if h == nil {
		h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "service initializing", http.StatusServiceUnavailable)
		})
	}
	a.v.Store(h)
	return a
}

func (a *AtomicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := a.v.Load().(http.Handler)
	h.ServeHTTP(w, r)
}

func (a *AtomicHandler) Swap(h http.Handler) {
	if h == nil {
		return
	}
	a.v.Store(h)
}
