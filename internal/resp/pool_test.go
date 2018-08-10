package resp_test

import (
	"testing"
	"time"

	"github.com/furui/gochunk/internal/resp"
	"github.com/furui/gochunk/pkg/mocks"
	"github.com/stretchr/testify/assert"
)

func TestPool(t *testing.T) {
	p := resp.NewPool(resp.NewConfig())
	err := p.Start()
	assert.NoError(t, err)
	c := mocks.NewMockConn()
	p.Queue(c)
	c.Write([]byte("!TEST\r\n"))
	time.Sleep(10 * time.Millisecond)
	// r := bufio.NewReader(c)
	// s, err := r.ReadString('\r')
	// assert.NoError(t, err)
	// assert.Equal(t, "-SomeError\r", s)
	err = p.Stop()
	assert.NoError(t, err)
}
