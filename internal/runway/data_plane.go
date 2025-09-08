package runway

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/nanthakumaran-s/golancer/internal/config"
)

type DataPlane struct {
	imm    *config.ImmutableConfig
	cfg    atomic.Value
	stopCh chan struct{}
}

func NewDataPlane(imm *config.ImmutableConfig) *DataPlane {
	return &DataPlane{
		imm:    imm,
		stopCh: make(chan struct{}),
	}
}

func (dp *DataPlane) UpdateConfig(cfg *config.MutableConfig) {
	dp.cfg.Store(cfg)
	fmt.Printf("DataPlane: hot config updated -> %+v\n", cfg)
}

func (dp *DataPlane) Start() {
	addr := fmt.Sprintf(":%d", dp.imm.Port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	fmt.Println("DataPlane: listening on", addr)

	go func() {
		defer ln.Close()

		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-dp.stopCh:
					return
				default:
					fmt.Println("Accept error:", err)
					continue
				}
			}

			go dp.handleConn(conn)
		}
	}()
}

func (d *DataPlane) handleConn(c net.Conn) {
	defer c.Close()
	buf := make([]byte, 1024)

	for {
		n, err := c.Read(buf)
		if err != nil {
			return
		}

		cfg := d.cfg.Load().(*config.MutableConfig)
		reply := fmt.Sprintf("[%s] echo: %s", cfg.Logging.Level, string(buf[:n]))
		c.Write([]byte(reply))
	}
}

func (dp *DataPlane) Stop() {
	close(dp.stopCh)
}
