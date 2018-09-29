package db

import (
	"path/filepath"
	"strconv"
	"sync"

	"github.com/etcd-io/bbolt"
	"github.com/furui/gochunk/pkg/config"
	"github.com/furui/gochunk/pkg/processor"
)

// Manager manages multiple databases
type Manager interface {
	Get(id int) (Database, error)
	Close() error
}

type manager struct {
	DB        *bbolt.DB
	proc      processor.Processor
	databases map[int]string
	mux       sync.Mutex
}

// NewManager creates a new database manager
func NewManager(conf *config.Config, proc processor.Processor) Manager {
	dbLocation := filepath.Join(conf.DatabaseLocation, "manager.db")
	DB, err := bbolt.Open(dbLocation, 0666, nil)
	if err != nil {
		panic(err)
	}
	m := &manager{
		DB:   DB,
		proc: proc,
	}
	err = m.load()
	if err != nil {
		panic(err)
	}
	return m
}

func (m *manager) load() error {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.databases = make(map[int]string)
	err := m.DB.View(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("id"))
		if err != nil {
			return err
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			key, err := strconv.Atoi(string(k))
			if err != nil {
				return err
			}
			m.databases[key] = string(v)
		}
		return nil
	})
	return err
}

func (m *manager) generate(id int) (string, error) {
	return "", nil
}

func (m *manager) Get(id int) (Database, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	var db string
	db, ok := m.databases[id]
	if !ok {
		uuid, err := m.generate(id)
		if err != nil {
			return nil, err
		}
		db = uuid
	}
	return NewDatabase(db), nil
}

func (m *manager) Close() error {
	return m.DB.Close()
}
