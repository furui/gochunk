package resp_test

import (
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/furui/gochunk/pkg/uuid"

	"github.com/furui/gochunk/pkg/config"
	"github.com/furui/gochunk/pkg/db"
	"github.com/furui/gochunk/pkg/processor"
	"github.com/furui/gochunk/pkg/state"
	respTypes "github.com/furui/gochunk/pkg/types"

	"github.com/furui/gochunk/pkg/mocks"
	"github.com/furui/gochunk/pkg/resp"
	"github.com/stretchr/testify/assert"
)

func init() {
	files, err := filepath.Glob(filepath.Join(os.TempDir(), "*.db"))
	if err != nil {
		return
	}
	for _, f := range files {
		err = os.Remove(f)
		if err != nil {
			log.Printf("warning: couldn't remove %s: %s", f, err)
		}
	}
}

func setupPool() (db.Manager, resp.Pool, *config.Config) {
	conf := config.NewConfig()
	conf.ReadTimeout = time.Second
	conf.DatabaseLocation = os.TempDir()

	data := db.NewManager(conf, uuid.NewGenerator())

	proc := processor.NewProcessor(data)
	proc.AddCommand("KEYS", func(dbManager db.Manager, state state.Client, params [][]byte) (respTypes.Type, error) {
		s := respTypes.SimpleString("TEST")
		return &respTypes.Array{Contents: []respTypes.Type{
			&s,
		}}, nil
	})

	p := resp.NewPool(conf, proc)
	return data, p, conf
}

func TestPool(t *testing.T) {
	data, p, _ := setupPool()
	defer data.Close()
	err := p.Start()
	assert.NoError(t, err)
	err = p.Start()
	assert.Error(t, err)
	err = p.Stop()
	assert.NoError(t, err)
}

func TestPoolResponses(t *testing.T) {
	data, p, conf := setupPool()
	defer data.Close()
	err := p.Start()
	assert.NoError(t, err)
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
		{
			desc:     "auth no passwd",
			write:    []byte("*2\r\n$4\r\nAUTH\r\n$4\r\ntest\r\n"),
			response: []byte("-Client sent AUTH, but no password set\r\n"),
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
			}
			assert.Equal(t, tC.response, slicebuf)
		})
	}

	conf.RequirePass = "test 1 2 3"
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
			write:    []byte("*2\r\n$4\r\nAUTH\r\n$10\r\ntest 1 2 3\r\n"),
			response: []byte("*1\r\n+OK\r\n"),
		},
		{
			desc:     "real command after auth",
			write:    []byte("*2\r\n$4\r\nAUTH\r\n$10\r\ntest 1 2 3\r\n*2\r\n$4\r\nKEYS\r\n$1\r\n*\r\n"),
			response: []byte("*1\r\n+OK\r\n*1\r\n+TEST\r\n"),
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
			}
			assert.Equal(t, tC.response, slicebuf)

		})
	}

	err = p.Stop()
	assert.NoError(t, err)
}

func TestIncompleteConnection(t *testing.T) {
	data, p, _ := setupPool()
	defer data.Close()
	err := p.Start()
	assert.NoError(t, err)
	s, c := mocks.NewMockConn()
	p.Queue(s)

	assert.NotPanics(t, func() {
		c.Write([]byte("*2\r\n$4\r\nKEYS\r\n$1\r\n*\r\n"))
		c.Close()

		err = p.Stop()
		assert.NoError(t, err)
	})

	data.Close()
}

func TestQuit(t *testing.T) {
	data, p, _ := setupPool()
	defer data.Close()
	err := p.Start()
	buf := make([]byte, 50)
	assert.NoError(t, err)
	s, c := mocks.NewMockConn()
	p.Queue(s)
	c.Write([]byte("*1\r\n$4\r\nQUIT\r\n"))
	n, err := c.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, []byte("+OK\r\n"), buf[:n])
	n, err = s.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	assert.Zero(t, n)
	assert.Error(t, err)
	err = p.Stop()
	assert.NoError(t, err)
}

func TestBuiltIns(t *testing.T) {
	data, p, _ := setupPool()
	defer data.Close()
	err := p.Start()
	assert.NoError(t, err)
	buf := make([]byte, 50)
	s, c := mocks.NewMockConn()
	p.Queue(s)

	testCases := []struct {
		desc     string
		write    []byte
		response []byte
	}{
		{
			desc:     "echo",
			write:    []byte("*2\r\n$4\r\nECHO\r\n$13\r\ntesting 1 2 3\r\n"),
			response: []byte("$13\r\ntesting 1 2 3\r\n"),
		},
		{
			desc:     "ping",
			write:    []byte("*1\r\n$4\r\nPING\r\n"),
			response: []byte("+PONG\r\n"),
		},
		{
			desc:     "ping payload",
			write:    []byte("*2\r\n$4\r\nPING\r\n$13\r\ntesting 1 2 3\r\n"),
			response: []byte("$13\r\ntesting 1 2 3\r\n"),
		},
		{
			desc:     "select",
			write:    []byte("*2\r\n$6\r\nSELECT\r\n$2\r\n11\r\n"),
			response: []byte("+OK\r\n"),
		},
		{
			desc:     "swap non existing db",
			write:    []byte("*3\r\n$6\r\nSWAPDB\r\n$2\r\n20\r\n$2\r\n30\r\n"),
			response: []byte("+OK\r\n"),
		},
		{
			desc:     "swap existing db",
			write:    []byte("*2\r\n$6\r\nSELECT\r\n$1\r\n5\r\n*2\r\n$6\r\nSELECT\r\n$1\r\n6\r\n*3\r\n$6\r\nSWAPDB\r\n$1\r\n5\r\n$1\r\n6\r\n"),
			response: []byte("+OK\r\n+OK\r\n+OK\r\n"),
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
			}
			assert.Equal(t, tC.response, slicebuf)
		})
	}

	err = p.Stop()
	assert.NoError(t, err)
}
