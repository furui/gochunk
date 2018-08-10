package resp

import (
	"fmt"
	"log"
	"net"
)

// Server provides a RESP server
type Server interface {
	Start() error
	Stop() error
}

type server struct {
	config   *Config
	listener net.Listener
	started  bool
	pool     Pool
}

func (s *server) Start() error {
	if s.started == false {
		err := s.pool.Start()
		if err != nil {
			return err
		}
		go func() {
			l, err := net.Listen("tcp", s.config.Host)
			if err != nil {
				panic(err)
			}
			defer l.Close()
			s.listener = l
			s.started = true
			for {
				conn, err := l.Accept()
				if err != nil {
					log.Fatal(err)
				}
				if s.started == false {
					break
				}
				s.pool.Queue(conn)
			}
		}()
	} else {
		return fmt.Errorf("Server already started")
	}
	return nil
}

func (s *server) Stop() error {
	if s.started == true {
		err := s.listener.Close()
		if err != nil {
			return err
		}
		err = s.pool.Stop()
		if err != nil {
			return err
		}
		s.started = false
	}
	return nil
}

// NewServer returns a new server
func NewServer(config *Config, pool Pool) Server {
	return &server{config: config, pool: pool, listener: nil, started: false}
}
