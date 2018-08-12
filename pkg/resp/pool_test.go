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
	c.Write([]byte("!TEST\r\n"))
	time.Sleep(10 * time.Millisecond)
	d := mocks.NewMockConn()
	p.Queue(d)
	err = p.Stop()
	assert.NoError(t, err)
}
