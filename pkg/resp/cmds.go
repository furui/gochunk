package resp

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/furui/gochunk/pkg/config"
	"github.com/furui/gochunk/pkg/db"
	"github.com/furui/gochunk/pkg/processor"
	"github.com/furui/gochunk/pkg/state"
	respTypes "github.com/furui/gochunk/pkg/types"
)

var (
	// ErrNoMatchingPass is thrown when auth fails
	ErrNoMatchingPass = errors.New("authentication required")
)

func addAuthCmd(config *config.Config, processor processor.Processor) {
	processor.AddCommand("AUTH", func(dbManager db.Manager, state state.Client, params [][]byte) (respTypes.Type, error) {
		if len(params) != 1 {
			return nil, fmt.Errorf("one parameter expected")
		}
		passwd := string(params[0])
		if ok, err := state.Authenticate(passwd); ok {
			if err != nil {
				return nil, err
			}
			s := respTypes.SimpleString("OK")
			return &respTypes.Array{Contents: []respTypes.Type{
				&s,
			}}, nil
		}
		return nil, ErrNoMatchingPass
	})
}

func addEchoCmd(config *config.Config, processor processor.Processor) {
	processor.AddCommand("ECHO", func(dbManager db.Manager, state state.Client, params [][]byte) (respTypes.Type, error) {
		if len(params) != 1 {
			return nil, fmt.Errorf("one parameter expected")
		}
		return &respTypes.BulkString{Data: params[0]}, nil
	})
}

func addPingCmd(config *config.Config, processor processor.Processor) {
	processor.AddCommand("PING", func(dbManager db.Manager, state state.Client, params [][]byte) (respTypes.Type, error) {
		if len(params) > 1 {
			return nil, fmt.Errorf("one or zero parameters expected")
		}
		if len(params) == 1 {
			return &respTypes.BulkString{Data: params[0]}, nil
		}
		t := respTypes.SimpleString("PONG")
		return &t, nil
	})
}

func addSelectCmd(config *config.Config, processor processor.Processor) {
	processor.AddCommand("SELECT", func(dbManager db.Manager, state state.Client, params [][]byte) (respTypes.Type, error) {
		if len(params) != 1 {
			return nil, fmt.Errorf("one parameter expected")
		}
		db, err := strconv.Atoi(string(params[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid index")
		}
		if db < 0 {
			return nil, fmt.Errorf("index out of range")
		}
		state.SetDatabase(db)
		t := respTypes.SimpleString("OK")
		return &t, nil
	})
}

func addQuitCmd(config *config.Config, processor processor.Processor) {
	processor.AddCommand("QUIT", func(dbManager db.Manager, state state.Client, params [][]byte) (respTypes.Type, error) {
		state.SetClosed(true)
		t := respTypes.SimpleString("OK")
		return &t, nil
	})
}

func addSwapDbCmd(config *config.Config, processor processor.Processor) {
	processor.AddCommand("SWAPDB", func(dbManager db.Manager, state state.Client, params [][]byte) (respTypes.Type, error) {
		if len(params) < 2 {
			return nil, fmt.Errorf("two parameters expected")
		}
		a, err := strconv.Atoi(string(params[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid index")
		}
		b, err := strconv.Atoi(string(params[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid index")
		}
		if (a < 0) || (b < 0) {
			return nil, fmt.Errorf("index out of range")
		}
		dbManager.Swap(a, b)
		t := respTypes.SimpleString("OK")
		return &t, nil
	})
}
