package main

import (
	"fmt"
	"log"

	"github.com/furui/gochunk/pkg/resp"
)

func main() {
	c := NewContainer()
	log.Print("Starting server")
	err := c.Invoke(func(s resp.Server) {
		err := s.Start()
		if err != nil {
			panic(err)
		}
	})
	if err != nil {
		panic(err)
	}
	log.Print("Server started")
	var s string
	fmt.Scanln(&s)
	err = c.Invoke(func(s resp.Server) {
		err := s.Stop()
		if err != nil {
			panic(err)
		}
	})
	if err != nil {
		panic(err)
	}
	log.Print("Server stopped")
}
