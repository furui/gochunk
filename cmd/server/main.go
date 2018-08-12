package main

import (
	"fmt"

	"github.com/furui/gochunk/pkg/resp"
)

func main() {
	c := NewContainer()
	c.Invoke(func(s resp.Server) {
		err := s.Start()
		if err != nil {
			panic(err)
		}
	})
	var s string
	fmt.Scan(&s)
	c.Invoke(func(s resp.Server) {
		err := s.Stop()
		if err != nil {
			panic(err)
		}
	})
}
