package main

import (
	"github.com/furui/gochunk/pkg/config"
	"github.com/furui/gochunk/pkg/db"
	"github.com/furui/gochunk/pkg/processor"
	"github.com/furui/gochunk/pkg/resp"
	"github.com/furui/gochunk/pkg/state"
	"github.com/furui/gochunk/pkg/uuid"
	"go.uber.org/dig"
)

// NewContainer returns a new DI container
func NewContainer() *dig.Container {
	c := dig.New()

	c.Provide(config.NewConfig)
	c.Provide(resp.NewPool)
	c.Provide(resp.NewScanner)
	c.Provide(resp.NewServer)
	c.Provide(processor.NewProcessor)
	c.Provide(db.NewManager)
	c.Provide(state.NewClient)
	c.Provide(uuid.NewGenerator)

	return c
}
