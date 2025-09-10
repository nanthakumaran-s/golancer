package tower

import (
	"context"
	"fmt"
	"time"

	"github.com/nanthakumaran-s/golancer/internal/config"
	"github.com/nanthakumaran-s/golancer/internal/proxy"
	"github.com/nanthakumaran-s/golancer/internal/runway"
	"github.com/nanthakumaran-s/golancer/internal/utils"
)

type ControlPlane struct {
	mailbox    chan interface{}
	done       chan struct{}
	dataplane  *runway.DataPlane
	cfgManager *config.Manager
	logger     *utils.Logger
}

func NewControlPlane(dp *runway.DataPlane) *ControlPlane {
	return &ControlPlane{
		mailbox:   make(chan interface{}, 16),
		done:      make(chan struct{}),
		dataplane: dp,
	}
}

func (cp *ControlPlane) Start(ctx context.Context) {
	logger, err := utils.NewLogger()
	if err != nil {
		panic(fmt.Sprintf("failed to init logger: %v", err))
	}
	logger.Start()
	cp.logger = logger
	cp.dataplane.SetLogger(logger)

	cp.logger.Info(utils.CONTROL_PLANE, "golancer logger Initialized")

	cfgManager, err := config.NewManager(cp.logger)
	if err != nil {
		cp.logger.Fatal("failed to init config manager: %v", err)
	}
	cp.cfgManager = cfgManager

	initial := cp.cfgManager.Get()
	if initial == nil {
		cp.logger.Fatal("failed to init config manager: %v", err)
	}

	rs, err := buildRouterState(cp.cfgManager.Get())
	if err != nil {
		cp.logger.Fatal("failed to build initial router state: %v", err)
	}
	cp.dataplane.UpdateHttpHandler(&proxy.Router{State: rs, Logger: cp.logger})

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
					cp.logger.Info(utils.CONTROL_PLANE, "apply config (from mailbox)")
					cp.dataplane.UpdateConfig(m.NewConfig)
					if rs, err := buildRouterState(m.NewConfig); err == nil {
						cp.dataplane.UpdateHttpHandler(&proxy.Router{State: rs, Logger: cp.logger})
						cp.logger.Info(utils.CONTROL_PLANE, "router state swapped")
					} else {
						cp.logger.Warn(utils.CONTROL_PLANE, fmt.Sprintln("buildRouterState error:", err))
					}

				case Shutdown:
					cp.logger.Info(utils.CONTROL_PLANE, "shutdown requested (from mailbox)")
					cp.dataplane.Stop()
					cp.logger.Stop()
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
				cp.logger.Info(utils.CONTROL_PLANE, "context cancelled, shutting down")
				cp.dataplane.Stop()
				cp.logger.Stop()
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
	tr := proxy.NewTransport(cfg.Proxy.MaxIdleConns, cfg.Proxy.IdleConnTimeout)
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
