package main

import (
	"github.com/furui/gochunk/pkg/config"
	"github.com/furui/gochunk/pkg/processor"
	"github.com/furui/gochunk/pkg/resp"
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

	return c
}
