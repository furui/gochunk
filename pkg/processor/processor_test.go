package processor

import (
	"errors"
	"testing"

	"github.com/furui/gochunk/pkg/db"
	"github.com/furui/gochunk/pkg/state"
	respTypes "github.com/furui/gochunk/pkg/types"
	"github.com/stretchr/testify/assert"
)

func processorDependencies() db.Manager {
	return nil
}

func TestAddExecuteCommand(t *testing.T) {
	p := NewProcessor(processorDependencies())
	testCases := []struct {
		desc     string
		command  string
		fn       Command
		data     [][]byte
		wantType respTypes.Type
		wantErr  error
	}{
		{
			desc:    "regular",
			command: "REGULAR",
			fn: func(dbManager db.Manager, state state.Client, params [][]byte) (respTypes.Type, error) {
				return nil, nil
			},
			data: [][]byte{
				[]byte("*"),
			},
			wantType: nil,
			wantErr:  nil,
		},
		{
			desc:    "fn returns err",
			command: "FNRETURNSERR",
			fn: func(dbManager db.Manager, state state.Client, params [][]byte) (respTypes.Type, error) {
				return nil, errors.New("test")
			},
			data: [][]byte{
				[]byte("*"),
			},
			wantType: nil,
			wantErr:  errors.New("test"),
		},
		{
			desc:    "fn returns type",
			command: "FNRETURNSTYPE",
			fn: func(dbManager db.Manager, state state.Client, params [][]byte) (respTypes.Type, error) {
				return &respTypes.Array{}, nil
			},
			data: [][]byte{
				[]byte("*"),
			},
			wantType: &respTypes.Array{},
			wantErr:  nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			state := state.NewClient()
			b := p.AddCommand(tC.command, tC.fn)
			assert.True(t, b)
			tp, err := p.Execute(tC.command, state, tC.data)
			assert.Equal(t, tC.wantErr, err)
			assert.Equal(t, tC.wantType, tp)
		})
	}
	t.Run("already added", func(t *testing.T) {
		added := p.AddCommand("REGULAR", func(dbManager db.Manager, state state.Client, params [][]byte) (respTypes.Type, error) {
			return nil, nil
		})
		assert.False(t, added)
	})
}

func TestDeleteCommand(t *testing.T) {
	p := NewProcessor(processorDependencies())
	added := p.AddCommand("REGULAR", func(dbManager db.Manager, state state.Client, params [][]byte) (respTypes.Type, error) {
		return nil, nil
	})
	assert.True(t, added)
	deleted := p.DeleteCommand("REGULA")
	assert.False(t, deleted)
	deleted = p.DeleteCommand("REGULAR")
	assert.True(t, deleted)
}

func TestExecuteNotFound(t *testing.T) {
	p := NewProcessor(processorDependencies())
	res, err := p.Execute("DOESNTEXIST", state.NewClient(), [][]byte{})
	assert.Error(t, err)
	assert.Nil(t, res)
}
