package server

import (
	"fmt"
	"net"
	"sync"

	"github.com/nanthakumaran-s/golancer/internal/config"
)

type Server struct {
	mu   sync.Mutex
	cfg  *config.Config
	addr string
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg:  cfg,
		addr: cfg.Server.Address,
	}
}

func (s *Server) Start() error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer l.Close()

	fmt.Printf("listening on %s\n", s.addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	msg := string(buf[:n])
	conn.Write([]byte("echo: " + msg))
}

func (s *Server) ApplyConfig(cfg *config.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = cfg
}
