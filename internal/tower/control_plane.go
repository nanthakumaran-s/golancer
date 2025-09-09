package tower

import (
	"context"
	"fmt"
	"time"

	"github.com/nanthakumaran-s/golancer/internal/config"
	"github.com/nanthakumaran-s/golancer/internal/proxy"
	"github.com/nanthakumaran-s/golancer/internal/runway"
)

type ControlPlane struct {
	mailbox    chan interface{}
	done       chan struct{}
	dataplane  *runway.DataPlane
	cfgManager *config.Manager
}

func NewControlPlane(dp *runway.DataPlane) *ControlPlane {
	return &ControlPlane{
		mailbox:   make(chan interface{}, 16),
		done:      make(chan struct{}),
		dataplane: dp,
	}
}

func (cp *ControlPlane) Start(ctx context.Context) {
	cfgManager, err := config.NewManager()
	if err != nil {
		panic(fmt.Sprintf("failed to init config manager: %v", err))
	}
	cp.cfgManager = cfgManager

	initial := cp.cfgManager.Get()
	if initial == nil {
		panic(fmt.Sprintf("failed to init config manager: %v", err))
	}
	cp.dataplane.UpdateConfig(initial)

	rs, err := buildRouterState(cp.cfgManager.Get())
	if err != nil {
		panic(fmt.Sprintf("failed to build initial router state: %v", err))
	}
	cp.dataplane.UpdateHttpHandler(&proxy.Router{State: rs})

	cp.dataplane.Start()

	updates := cp.cfgManager.Subscribe()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case msg := <-cp.mailbox:
				switch m := msg.(type) {
				case ConfigUpdated:
					if m.NewConfig == nil {
						continue
					}
					fmt.Println("[control] apply config (from mailbox)")
					cp.dataplane.UpdateConfig(m.NewConfig)
					if rs, err := buildRouterState(m.NewConfig); err == nil {
						cp.dataplane.UpdateHttpHandler(&proxy.Router{State: rs})
						fmt.Println("[control] router state swapped")
					} else {
						fmt.Println("[control] buildRouterState error:", err)
					}

				case Shutdown:
					fmt.Println("[control] Shutdown requested")
					cp.dataplane.Stop()
					close(cp.done)
					return
				}

			case <-ticker.C:
				continue

			case cfg := <-updates:
				cp.mailbox <- ConfigUpdated{
					NewConfig: cfg,
				}

			case <-ctx.Done():
				fmt.Println("[control] context cancelled, shutting down")
				cp.dataplane.Stop()
				close(cp.done)
				return
			}
		}
	}()
}

func (cp *ControlPlane) Send(msg interface{}) {
	cp.mailbox <- msg
}

func (cp *ControlPlane) Stop() {
	cp.mailbox <- Shutdown{}
	<-cp.done
}

func buildRouterState(cfg *config.Config) (*proxy.RouterState, error) {
	tr := proxy.NewTransport(cfg.Proxy.MaxIdleConnections, cfg.Proxy.IdleConnectionTimeout)
	rs := &proxy.RouterState{
		Transport: tr,
		Timeout:   cfg.Proxy.DefaultTimeout,
	}
	rs.Routes = make([]proxy.Route, 0, len(cfg.Routes))
	for _, r := range cfg.Routes {
		pool, err := proxy.NewUpstreamPool(r.Upstreams)
		if err != nil {
			return nil, err
		}
		rs.Routes = append(rs.Routes, proxy.Route{
			Name:       r.Name,
			Hosts:      r.Match.Hosts,
			PathPrefix: r.Match.PathPrefix,
			Pool:       pool,
		})
	}
	return rs, nil
}
