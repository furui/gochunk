package resp

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// Pool contains a thread pool that handles connections
type Pool interface {
	Queue(conn net.Conn)
	Start() error
	Stop() error
}

type pool struct {
	connections  []net.Conn
	threads      int
	started      bool
	mutex        *sync.Mutex
	readTimeout  time.Duration
	writeTimeout time.Duration
	config       *Config
}

func (p *pool) Queue(conn net.Conn) {
	p.mutex.Lock()
	p.connections = append(p.connections, conn)
	p.mutex.Unlock()
}

func (p *pool) dequeue() net.Conn {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if len(p.connections) == 0 {
		return nil
	}
	c := p.connections[0]
	p.connections = p.connections[1:]
	return c
}

func (p *pool) Start() error {
	if p.started == true {
		return fmt.Errorf("Pool already started")
	}
	p.started = true
	for i := 0; i < p.threads; i++ {
		go p.thread()
	}
	return nil
}

func (p *pool) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.started = false
	return p.kill()
}

func (p *pool) kill() error {
	for _, c := range p.connections {
		err := c.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *pool) thread() {
	for p.started != false {
		conn := p.dequeue()
		if conn == nil {
			continue
		}
		scanner := NewScanner(conn)
		writer := bufio.NewWriter(conn)
		for conn.SetReadDeadline(time.Now().Add(p.readTimeout)); scanner.Scan(); conn.SetReadDeadline(time.Now().Add(p.readTimeout)) {
			conn.SetWriteDeadline(time.Now().Add(p.writeTimeout))
			if err := scanner.Err(); err != nil {
				e := Error(fmt.Sprintf("Scan Error: %s", err))
				log.Printf("Scan erorr %s: %s", conn.RemoteAddr().String(), err)
				b := e.Bytes()
				_, err = writer.Write(b)
				if err != nil {
					log.Printf("Couldn't write to %s: %s", conn.RemoteAddr().String(), err)
				}
				err = writer.Flush()
				if err != nil {
					log.Printf("Couldn't flush buffer to %s: %s", conn.RemoteAddr().String(), err)
				}
			} else {
				t := scanner.Type()
				log.Printf("%+v", t.Value())
			}
		}
		conn.Close()
	}
}

// NewPool creates a new thread pool
func NewPool(config *Config) Pool {
	p := &pool{
		connections:  make([]net.Conn, 0),
		threads:      config.Workers,
		started:      false,
		mutex:        &sync.Mutex{},
		readTimeout:  config.ReadTimeout,
		writeTimeout: config.WriteTimeout,
	}
	return p
}
