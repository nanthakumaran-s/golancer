package runway

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nanthakumaran-s/golancer/internal/config"
	"github.com/nanthakumaran-s/golancer/internal/proxy"
)

type DataPlane struct {
	srv         *config.ServerDefaults
	cfg         atomic.Value
	httpHandler *proxy.AtomicHandler
	listener    net.Listener
	server      *http.Server
	stopOnce    sync.Once
	stopCh      chan struct{}
}

func NewDataPlane(srv *config.ServerDefaults) *DataPlane {
	return &DataPlane{
		srv:    srv,
		stopCh: make(chan struct{}),
	}
}

func (dp *DataPlane) UpdateConfig(cfg *config.Config) {
	if cfg == nil {
		return
	}
	dp.cfg.Store(cfg)
	fmt.Printf("[dataplane] hot config updated -> %+v\n", cfg)
}

func (dp *DataPlane) UpdateHttpHandler(h http.Handler) {
	if h == nil {
		return
	}

	if dp.httpHandler == nil {
		dp.httpHandler = proxy.NewAtomicHandler(h)
		return
	}

	dp.httpHandler.Swap(h)
}

func (dp *DataPlane) Start() {
	addr := fmt.Sprintf(":%d", dp.srv.Port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		panic(fmt.Sprintf("[dataplane] failed to bind %s: %v", addr, err))
	}
	dp.listener = ln

	if dp.httpHandler == nil {
		dp.httpHandler = proxy.NewAtomicHandler(nil)
	}

	dp.server = &http.Server{
		Handler:           dp.httpHandler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	fmt.Println("[dataplane] listening on", addr)

	go func() {
		if err := dp.server.Serve(ln); err != nil && err != http.ErrServerClosed {
			fmt.Println("[dataplane] http server error:", err)
		}
	}()
}

func (dp *DataPlane) Stop() {
	dp.stopOnce.Do(func() {
		close(dp.stopCh)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if dp.server != nil {
			if err := dp.server.Shutdown(ctx); err != nil {
				fmt.Println("[dataplane] graceful shutdown error:", err)
			}
		}

		if dp.listener != nil {
			_ = dp.listener.Close()
		}
	})
}
