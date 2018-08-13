package resp

import (
	"runtime"
	"time"
)

// Config defines the configuration for the RESP server
type Config struct {
	Host         string
	Workers      int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewConfig reads a new config
func NewConfig() *Config {
	return &Config{
		Host:         "127.0.0.1:3030",
		Workers:      runtime.NumCPU(),
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}
}
