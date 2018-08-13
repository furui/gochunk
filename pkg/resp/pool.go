package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/furui/gochunk/pkg/processor"
	respTypes "github.com/furui/gochunk/pkg/types"
)

var (
	// ErrInvalidType is thrown if the server doesn't receive an array of bulkstrings
	ErrInvalidType = errors.New("received invalid type")
	// ErrScan is thrown if the server receives invalid data
	ErrScan = errors.New("scan error")
	// ErrEmptyArray is thrown when an empty array is sent
	ErrEmptyArray = errors.New("received empty array")
	// ErrInvalidData is thrown when the array doesn't contain all bulk strings
	ErrInvalidData = errors.New("received invalid data")
	// ErrWrite is thrown when a write fails
	ErrWrite = errors.New("write failed")
	// ErrFlush is thrown when a flush fails
	ErrFlush = errors.New("flush failed")
)

// Pool contains a thread pool that handles connections
type Pool interface {
	Queue(conn net.Conn)
	Start() error
	Stop() error
}

type pool struct {
	processor    processor.Processor
	connections  []net.Conn
	threads      int
	started      bool
	mutex        *sync.Mutex
	cond         *sync.Cond
	readTimeout  time.Duration
	writeTimeout time.Duration
	config       *Config
}

func (p *pool) Lock() {
	p.mutex.Lock()
}

func (p *pool) Unlock() {
	p.mutex.Unlock()
}

func (p *pool) Queue(conn net.Conn) {
	p.Lock()
	p.connections = append(p.connections, conn)
	p.Unlock()
	p.cond.Signal()
}

func (p *pool) dequeue() net.Conn {
	p.Lock()
	defer p.Unlock()
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
		p.Lock()
		p.cond.Wait()
		p.Unlock()
		conn := p.dequeue()
		if conn == nil {
			continue
		}
		scanner := NewScanner(conn)
		writer := bufio.NewWriter(conn)
		for conn.SetReadDeadline(time.Now().Add(p.readTimeout)); scanner.Scan(); conn.SetReadDeadline(time.Now().Add(p.readTimeout)) {
			conn.SetWriteDeadline(time.Now().Add(p.writeTimeout))
			if err := scanner.Err(); err != nil {
				e := sendError(writer, ErrScan.Error())
				log.Printf("scan error %s: %s", conn.RemoteAddr().String(), err)
				if e != nil {
					log.Printf("couldn't send error to %s: %s", conn.RemoteAddr().String(), e)
				}
				if e == io.EOF || e == io.ErrClosedPipe || e == io.ErrUnexpectedEOF {
					break
				}
				continue
			}
			res, ok := scanner.Type().(*respTypes.Array)
			if !ok {
				e := sendError(writer, ErrInvalidType.Error())
				log.Printf("invalid type %s", conn.RemoteAddr().String())
				if e != nil {
					log.Printf("couldn't send error to %s: %s", conn.RemoteAddr().String(), e)
				}
				if e == io.EOF || e == io.ErrClosedPipe || e == io.ErrUnexpectedEOF {
					break
				}
				continue
			}
			if len(res.Contents) < 1 {
				e := sendError(writer, ErrEmptyArray.Error())
				log.Printf("empty array %s", conn.RemoteAddr().String())
				if e != nil {
					log.Printf("couldn't send error to %s: %s", conn.RemoteAddr().String(), e)
				}
				if e == io.EOF || e == io.ErrClosedPipe || e == io.ErrUnexpectedEOF {
					break
				}
				continue
			}
			if !containsAllBulkStrings(res.Contents) {
				e := sendError(writer, ErrInvalidData.Error())
				log.Printf("invalid data %s", conn.RemoteAddr().String())
				if e != nil {
					log.Printf("couldn't send error to %s: %s", conn.RemoteAddr().String(), e)
				}
				if e == io.EOF || e == io.ErrClosedPipe || e == io.ErrUnexpectedEOF {
					break
				}
				continue
			}
			cmd := string(res.Contents[0].Value().([]byte))
			params := [][]byte{}
			for _, v := range res.Contents[1:] {
				params = append(params, v.Value().([]byte))
			}
			response, err := p.processor.Execute(cmd, params)
			if err != nil {
				e := sendError(writer, err.Error())
				if e != nil {
					log.Printf("couldn't send error to %s: %s", conn.RemoteAddr().String(), e)
				}
				if e == io.EOF || e == io.ErrClosedPipe || e == io.ErrUnexpectedEOF {
					break
				}
				continue
			}
			if _, err := response.Stream(writer); err != nil {
				if err == io.EOF || err == io.ErrClosedPipe || err == io.ErrUnexpectedEOF {
					break
				}
				log.Printf("couldn't stream to %s: %s", conn.RemoteAddr().String(), err)
			}
			if err := writer.Flush(); err != nil {
				if err == io.EOF || err == io.ErrClosedPipe || err == io.ErrUnexpectedEOF {
					break
				}
				log.Printf("couldn't flush to %s: %s", conn.RemoteAddr().String(), err)
			}
		}
		conn.Close()
	}
}

// NewPool creates a new thread pool
func NewPool(config *Config, processor processor.Processor) Pool {
	p := &pool{
		processor:    processor,
		connections:  make([]net.Conn, 0),
		threads:      config.Workers,
		started:      false,
		mutex:        &sync.Mutex{},
		readTimeout:  config.ReadTimeout,
		writeTimeout: config.WriteTimeout,
	}
	p.cond = sync.NewCond(p)
	return p
}

func sendError(w *bufio.Writer, msg string) error {
	e := respTypes.Error(msg)
	b := e.Bytes()
	_, err := w.Write(b)
	if err != nil {
		return ErrWrite
	}
	err = w.Flush()
	if err != nil {
		return ErrFlush
	}
	return nil
}

func containsAllBulkStrings(rt []respTypes.Type) bool {
	for _, t := range rt {
		_, ok := t.(*respTypes.BulkString)
		if !ok {
			return false
		}
	}
	return true
}
