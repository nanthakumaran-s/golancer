package server

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/nanthakumaran-s/golancer/internal/config"
)

type Server struct {
	mu   sync.RWMutex
	cfg  *config.Config
	addr string
}

func NewServer(addr string, cfg *config.Config) *Server {
	return &Server{
		addr: addr,
		cfg:  cfg,
	}
}

func (s *Server) Start() error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer l.Close()

	fmt.Printf("[server] listening on %s\n", s.addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(c net.Conn) {
	defer c.Close()
	buf := make([]byte, 1024)

	for {
		n, err := c.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Println("read error:", err)
			}
			return
		}

		msg := string(buf[:n])

		s.mu.RLock()
		reply := fmt.Sprintf("[%s] echo: %s", s.cfg.Logging.Level, msg)
		s.mu.RUnlock()

		c.Write([]byte(reply))
	}
}

func (s *Server) ApplyConfig(cfg *config.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = cfg
}
