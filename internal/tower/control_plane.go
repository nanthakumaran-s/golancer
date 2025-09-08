package tower

import (
	"context"
	"fmt"
	"time"

	"github.com/nanthakumaran-s/golancer/internal/config"
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
	cp.dataplane.UpdateConfig(cfgManager.Get())
	cp.cfgManager = cfgManager

	updates := cp.cfgManager.Subscribe()
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case msg := <-cp.mailbox:
				switch m := msg.(type) {
				case ConfigUpdated:
					fmt.Printf("Config updated: %+v\n", m.NewConfig)
					cp.dataplane.UpdateConfig(m.NewConfig)

				case Shutdown:
					fmt.Println("Control plane shutting down...")
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
				close(cp.done)
				return
			}
		}
	}()

	cp.dataplane.Start()
}

func (cp *ControlPlane) Send(msg interface{}) {
	cp.mailbox <- msg
}

func (cp *ControlPlane) Stop() {
	cp.mailbox <- Shutdown{}
	<-cp.done
}
