package resp_test

import (
	"log"
	"testing"

	"github.com/furui/gochunk/pkg/config"
	"github.com/furui/gochunk/pkg/processor"
	respTypes "github.com/furui/gochunk/pkg/types"

	"github.com/furui/gochunk/pkg/mocks"
	"github.com/furui/gochunk/pkg/resp"
	"github.com/stretchr/testify/assert"
)

func TestPoolResponses(t *testing.T) {
	proc := processor.NewProcessor()
	proc.AddCommand("KEYS", func(params [][]byte) (respTypes.Type, error) {
		s := respTypes.SimpleString("TEST")
		return &respTypes.Array{Contents: []respTypes.Type{
			&s,
		}}, nil
	})
	conf := config.NewConfig()
	p := resp.NewPool(conf, proc)
	err := p.Start()
	assert.NoError(t, err)
	err = p.Start()
	assert.Error(t, err)
	buf := make([]byte, 50)
	s, c := mocks.NewMockConn()
	p.Queue(s)

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
		{
			desc:     "multi command",
			write:    []byte("*2\r\n$4\r\nKEYS\r\n$1\r\n*\r\n*2\r\n$4\r\nKEYS\r\n$1\r\n*\r\n"),
			response: []byte("*1\r\n+TEST\r\n*1\r\n+TEST\r\n"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			c.Write(tC.write)
			slicebuf := make([]byte, 0)
			for len(slicebuf) < len(tC.response) {
				n, err := c.Read(buf)
				assert.NoError(t, err)
				slicebuf = append(slicebuf, buf[:n]...)
				log.Printf(string(buf[:n]))
			}
			assert.Equal(t, tC.response, slicebuf)
		})
	}

	conf.RequirePass = "test"
	s, c = mocks.NewMockConn()
	p.Queue(s)
	authCases := []struct {
		desc     string
		write    []byte
		response []byte
	}{
		{
			desc:     "noauth",
			write:    []byte("*2\r\n$4\r\nKEYS\r\n$1\r\n*\r\n"),
			response: []byte("-authentication required\r\n"),
		},
		{
			desc:     "incorrect auth",
			write:    []byte("*2\r\n$4\r\nAUTH\r\n$4\r\nnope\r\n"),
			response: []byte("-authentication required\r\n"),
		},
		{
			desc:     "correct auth",
			write:    []byte("*2\r\n$4\r\nAUTH\r\n$4\r\ntest\r\n"),
			response: []byte("*1\r\n+OK\r\n"),
		},
		{
			desc:     "real command after auth",
			write:    []byte("*2\r\n$4\r\nKEYS\r\n$1\r\n*\r\n"),
			response: []byte("*1\r\n+TEST\r\n"),
		},
	}
	for _, tC := range authCases {
		t.Run(tC.desc, func(t *testing.T) {
			c.Write(tC.write)
			slicebuf := make([]byte, 0)
			for len(slicebuf) < len(tC.response) {
				n, err := c.Read(buf)
				assert.NoError(t, err)
				slicebuf = append(slicebuf, buf[:n]...)
				log.Printf(string(buf[:n]))
			}
			assert.Equal(t, tC.response, slicebuf)

		})
	}

	err = p.Stop()
	assert.NoError(t, err)
}
