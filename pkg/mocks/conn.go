package mocks

import (
	"io"
	"net"
	"time"
)

type MockConn struct {
	net.Conn
	ReadPipeReader  *io.PipeReader
	ReadPipeWriter  *io.PipeWriter
	WritePipeReader *io.PipeReader
	WritePipeWriter *io.PipeWriter
	closed          bool
}

func (c *MockConn) Read(b []byte) (n int, err error) {
	if c.closed {
		return 0, io.EOF
	}
	return c.ReadPipeReader.Read(b)
}

func (c *MockConn) Write(b []byte) (n int, err error) {
	if c.closed {
		return 0, io.EOF
	}
	return c.WritePipeWriter.Write(b)
}

func (c *MockConn) Close() error {
	c.closed = true
	return nil
}

func (c *MockConn) LocalAddr() net.Addr {
	return &net.IPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Zone: "",
	}
}

func (c *MockConn) RemoteAddr() net.Addr {
	return &net.IPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Zone: "",
	}
}

func (c *MockConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// NewMockConn returns a new Mock net.Conn interface
func NewMockConn() *MockConn {
	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()
	return &MockConn{
		ReadPipeReader:  r1,
		ReadPipeWriter:  w1,
		WritePipeReader: r2,
		WritePipeWriter: w2,
		closed:          false,
	}
}
