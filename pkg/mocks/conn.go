package mocks

import (
	"bytes"
	"io"
	"net"
	"time"
)

type mockConn struct {
	net.Conn
	buffer *bytes.Buffer
	closed bool
}

func (c *mockConn) Read(b []byte) (n int, err error) {
	if c.closed {
		return 0, io.EOF
	}
	return c.buffer.Read(b)
}

func (c *mockConn) Write(b []byte) (n int, err error) {
	if c.closed {
		return 0, io.EOF
	}
	return c.buffer.Write(b)
}

func (c *mockConn) Close() error {
	c.closed = true
	return nil
}

func (c *mockConn) LocalAddr() net.Addr {
	return &net.IPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Zone: "",
	}
}

func (c *mockConn) RemoteAddr() net.Addr {
	return &net.IPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Zone: "",
	}
}

func (c *mockConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *mockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// NewMockConn returns a new Mock net.Conn interface
func NewMockConn() net.Conn {
	return &mockConn{
		buffer: bytes.NewBufferString(""),
		closed: false,
	}
}
