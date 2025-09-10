package runway

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/nanthakumaran-s/golancer/internal/config"
	"github.com/nanthakumaran-s/golancer/internal/proxy"
	"github.com/nanthakumaran-s/golancer/internal/utils"
)

type DataPlane struct {
	srv         *config.ServerDefaults
	cfg         atomic.Value
	httpHandler *proxy.AtomicHandler
	logger      *utils.Logger
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
	dp.logger.Info(utils.DATA_PLANE, fmt.Sprintf("hot config updated -> %+v\n", cfg))
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

func (dp *DataPlane) SetLogger(lg *utils.Logger) {
	if dp.logger != nil {
		return
	}

	dp.logger = lg
}

func (dp *DataPlane) Start() {
	addr := fmt.Sprintf(":%d", dp.srv.Port)

	server := &http.Server{
		Addr:    addr,
		Handler: dp.httpHandler,
	}

	go func() {
		if dp.srv.UseTLS {
			if dp.srv.Local {
				cert, err := utils.GenerateSelfSignedCert()
				if err != nil {
					panic(err)
				}

				server.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
				dp.logger.Info(utils.DATA_PLANE, fmt.Sprintln("listening with self-signed TLS on", addr))
				server.ListenAndServeTLS("", "")
			} else {
				// TODO: AutoCert implementation
			}
		} else {
			dp.logger.Info(utils.DATA_PLANE, fmt.Sprintln("listening on", addr))
			server.ListenAndServe()
		}
	}()
}

func (dp *DataPlane) Stop() {
	close(dp.stopCh)
}
