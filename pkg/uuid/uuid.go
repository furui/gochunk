package uuid

import (
	"fmt"
	"sync"
	"time"
)

// Generator returns UUIDs
type Generator interface {
	GenerateTimeCounter() UUID
	Reset()
}

type generator struct {
	counter uint64
	mux     sync.Mutex
}

// UUID is a generated UUID
type UUID interface {
	Version() int
	Variant() int
	String() string
}

type uuidTimeCounter struct {
	low         int64
	mid         int64
	high        int64
	lowCounter  uint64
	highCounter uint64
}

// NewGenerator creates a UUID generator
func NewGenerator() Generator {
	g := &generator{}
	g.Reset()
	return g
}

func (g *generator) Reset() {
	g.counter = 0
}

// GenerateTimeCounter creates a non-cryptographic UUID based off the current time and a counter
func (g *generator) GenerateTimeCounter() UUID {
	t := time.Now()
	n := t.UnixNano()
	low := n & 0xFFFFFFFF
	n = n >> 32
	mid := n & 0xFFFF
	n = n >> 16
	high := n & 0xFFF
	g.mux.Lock()
	lowCounter := g.counter & 0xFFFFFFFFFFFF
	highCounter := (g.counter >> 48) & 0xFFF
	g.counter++
	g.mux.Unlock()
	return &uuidTimeCounter{
		low:         low,
		mid:         mid,
		high:        high,
		lowCounter:  lowCounter,
		highCounter: highCounter,
	}
}

func (u *uuidTimeCounter) Version() int {
	return 2
}

func (u *uuidTimeCounter) Variant() int {
	return 0
}

func (u *uuidTimeCounter) String() string {
	uuid := fmt.Sprintf("%08x-%04x-2%03x-0%03x-%012x", u.low, u.mid, u.high, u.highCounter, u.lowCounter)
	return uuid
}
