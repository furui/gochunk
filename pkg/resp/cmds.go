package resp

import (
	"fmt"

	"github.com/furui/gochunk/pkg/config"
	"github.com/furui/gochunk/pkg/processor"
	respTypes "github.com/furui/gochunk/pkg/types"
)

func addAuthCmd(config *config.Config, processor processor.Processor) {
	processor.AddCommand("AUTH", func(params [][]byte) (respTypes.Type, error) {
		if len(params) != 1 {
			return nil, fmt.Errorf("one parameter expected")
		}
		passwd := string(params[0])
		if passwd == config.RequirePass {
			s := respTypes.SimpleString("OK")
			return &respTypes.Array{Contents: []respTypes.Type{
				&s,
			}}, nil
		}
		return nil, ErrNoAuth
	})
}

func addEchoCmd(config *config.Config, processor processor.Processor) {
	processor.AddCommand("ECHO", func(params [][]byte) (respTypes.Type, error) {
		if len(params) != 1 {
			return nil, fmt.Errorf("one parameter expected")
		}
		return &respTypes.BulkString{Data: params[0]}, nil
	})
}

func addPingCmd(config *config.Config, processor processor.Processor) {
	processor.AddCommand("PING", func(params [][]byte) (respTypes.Type, error) {
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
