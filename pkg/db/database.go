package db

import (
	"fmt"
	"path/filepath"

	bbolt "github.com/etcd-io/bbolt"
	"github.com/furui/gochunk/pkg/config"
)

// Database is a structure for accessing a database
type Database interface {
	Close() error
}

type database struct {
	DB   *bbolt.DB
	conf *config.Config
}

// NewDatabase returns a database
func NewDatabase(filename string, conf *config.Config) Database {
	dbLocation := filepath.Join(conf.DatabaseLocation, fmt.Sprintf("%s.db", filename))
	DB, err := bbolt.Open(dbLocation, 0666, nil)
	if err != nil {
		panic(err)
	}
	return &database{
		DB:   DB,
		conf: conf,
	}
}

func (d *database) Close() error {
	return d.DB.Close()
}
