package mocks

import (
	"container/list"
	"fmt"
	"net"
	"sync"
)

type Conn struct {
	net.Conn
	Other  *Conn
	Reader *list.List
	Mux    sync.RWMutex
	closed bool
}

func freeport() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port, nil
}

// NewMockConn returns a set of server and client mock connections
func NewMockConn() (net.Conn, net.Conn) {
	freeport, err := freeport()
	if err != nil {
		panic(err)
	}
	ln, _ := net.Listen("tcp", fmt.Sprintf(":%d", freeport))
	client := make(chan net.Conn, 0)
	go func(client chan net.Conn) {
		t, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", freeport))
		if err == nil {
			goto DONE
		}
		for err != nil {
			t, err = net.Dial("tcp", fmt.Sprintf("localhost:%d", freeport))
		}
	DONE:
		client <- t
	}(client)
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		return conn, <-client
	}
}
