package processor

import (
	"fmt"

	"github.com/armon/go-radix"
	respTypes "github.com/furui/gochunk/pkg/types"
)

// Command is executed when a matching command is matched
type Command func(params [][]byte) (respTypes.Type, error)

// Processor executes commands passed in from the network connection
type Processor interface {
	AddCommand(command string, fn Command) bool
	DeleteCommand(command string) bool
	Execute(command string, params [][]byte) (respTypes.Type, error)
}

type processor struct {
	r *radix.Tree
}

// NewProcessor creates a new Processor interface
func NewProcessor() Processor {
	return &processor{
		r: radix.New(),
	}
}

func (p *processor) AddCommand(command string, fn Command) bool {
	_, exist := p.r.Get(command)
	if exist {
		return false
	}
	p.r.Insert(command, fn)
	return true
}

func (p *processor) DeleteCommand(command string) bool {
	_, deleted := p.r.Delete(command)
	return deleted
}

func (p *processor) Execute(command string, params [][]byte) (respTypes.Type, error) {
	c, exist := p.r.Get(command)
	if !exist {
		return nil, fmt.Errorf("unknown command '%s'", command)
	}
	return c.(Command)(params)
}
