package resp_test

import (
	"testing"
	"time"

	"github.com/furui/gochunk/pkg/config"
	"github.com/furui/gochunk/pkg/processor"
	respTypes "github.com/furui/gochunk/pkg/types"

	"github.com/furui/gochunk/pkg/mocks"
	"github.com/furui/gochunk/pkg/resp"
	"github.com/stretchr/testify/assert"
)

func TestPool(t *testing.T) {
	p := resp.NewPool(config.NewConfig(), processor.NewProcessor())
	err := p.Start()
	assert.NoError(t, err)
	err = p.Start()
	assert.Error(t, err)
	c := mocks.NewMockConn()
	p.Queue(c)

	d := mocks.NewMockConn()
	p.Queue(d)
	err = p.Stop()
	assert.NoError(t, err)
}

func TestPoolResponses(t *testing.T) {
	proc := processor.NewProcessor()
	proc.AddCommand("KEYS", func(params [][]byte) (respTypes.Type, error) {
		s := respTypes.SimpleString("TEST")
		return &respTypes.Array{Contents: []respTypes.Type{
			&s,
		}}, nil
	})
	p := resp.NewPool(config.NewConfig(), proc)
	err := p.Start()
	assert.NoError(t, err)
	err = p.Start()
	assert.Error(t, err)
	c := mocks.NewMockConn()
	p.Queue(c)
	buf := make([]byte, 50)

	testCases := []struct {
		desc     string
		write    []byte
		response []byte
	}{
		{
			desc:     "scan error",
			write:    []byte("!TEST\r\n"),
			response: []byte("-unknown command '!TEST'\r\n"),
		},
		{
			desc:     "command not found",
			write:    []byte("*1\r\n$4\r\nKAYS\r\n"),
			response: []byte("-unknown command 'KAYS'\r\n"),
		},
		{
			desc:     "not array",
			write:    []byte("+FOO\r\n"),
			response: []byte("-received invalid type\r\n"),
		},
		{
			desc:     "empty array",
			write:    []byte("*0\r\n"),
			response: []byte("-received empty array\r\n"),
		},
		{
			desc:     "not all bulk strings",
			write:    []byte("*2\r\n$4\r\nKEYS\r\n+FOO\r\n"),
			response: []byte("-received invalid data\r\n"),
		},
		{
			desc:     "real command",
			write:    []byte("*2\r\n$4\r\nKEYS\r\n$1\r\n*\r\n"),
			response: []byte("*1\r\n+TEST\r\n"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			c.ReadPipeWriter.Write(tC.write)
			time.Sleep(100 * time.Millisecond)
			n, err := c.WritePipeReader.Read(buf)
			assert.NoError(t, err)
			assert.Equal(t, tC.response, buf[:n])
		})
	}

	err = p.Stop()
	assert.NoError(t, err)
}
