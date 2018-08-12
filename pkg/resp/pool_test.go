package resp_test

import (
	"testing"
	"time"

	"github.com/furui/gochunk/pkg/processor"

	"github.com/furui/gochunk/pkg/mocks"
	"github.com/furui/gochunk/pkg/resp"
	"github.com/stretchr/testify/assert"
)

func TestPool(t *testing.T) {
	p := resp.NewPool(resp.NewConfig(), processor.NewProcessor())
	err := p.Start()
	assert.NoError(t, err)
	err = p.Start()
	assert.Error(t, err)
	c := mocks.NewMockConn()
	p.Queue(c)

	buf := make([]byte, 50)

	c.ReadPipeWriter.Write([]byte("!TEST\r\n"))
	time.Sleep(100 * time.Millisecond)
	n, err := c.WritePipeReader.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, "-scan error\r\n", string(buf[:n]))

	c.ReadPipeWriter.Write([]byte("*1\r\n$4\r\nKAYS\r\n"))
	time.Sleep(100 * time.Millisecond)
	n, err = c.WritePipeReader.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, "-command not found\r\n", string(buf[:n]))

	d := mocks.NewMockConn()
	p.Queue(d)
	err = p.Stop()
	assert.NoError(t, err)
}
