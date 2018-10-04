package db

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"sync"

	bbolt "github.com/etcd-io/bbolt"
	"github.com/furui/gochunk/pkg/uuid"

	"github.com/furui/gochunk/pkg/config"
)

// Manager manages multiple databases
type Manager interface {
	Swap(a int, b int) error
	Get(id int) (Database, error)
	Close() error
}

type manager struct {
	DB        *bbolt.DB
	uuid      uuid.Generator
	databases map[int]string
	pool      map[string]Database
	mux       sync.Mutex
	closed    bool
}

var (
	// ErrorManagerClosed returned when the manager is closed
	ErrorManagerClosed = fmt.Errorf("manager is closed")
	// ErrorFirstIndexNonExistant returned when the first index doesn't exist
	ErrorFirstIndexNonExistant = fmt.Errorf("first index non-existant")
	// ErrorSecondIndexNonExistant returned when the second index doesn't exist
	ErrorSecondIndexNonExistant = fmt.Errorf("second index non-existant")
)

// NewManager creates a new database manager
func NewManager(conf *config.Config, uuid uuid.Generator) Manager {
	dbLocation := filepath.Join(conf.DatabaseLocation, "manager.db")
	DB, err := bbolt.Open(dbLocation, 0666, nil)
	if err != nil {
		panic(err)
	}
	m := &manager{
		DB:     DB,
		uuid:   uuid,
		pool:   make(map[string]Database),
		closed: false,
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
		b := tx.Bucket([]byte("id"))
		if b == nil {
			return nil
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

func (m *manager) save() {
	defer func() {
		if x := recover(); x != nil {
			log.Printf("panic during db manager save: %v", x)
		}
	}()
	err := m.DB.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("id"))
		if err != nil {
			return err
		}
		for k, v := range m.databases {
			key := []byte(strconv.Itoa(k))
			val := b.Get(key)
			if val == nil || string(val) != v {
				err := b.Put(key, []byte(v))
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func (m *manager) generate(id int) string {
	u := m.uuid.GenerateTimeCounter()
	m.databases[id] = u.String()
	m.save()
	return m.databases[id]
}

func (m *manager) Swap(a int, b int) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	aSide, ok := m.databases[a]
	if !ok {
		return ErrorFirstIndexNonExistant
	}
	bSide, ok := m.databases[b]
	if !ok {
		return ErrorSecondIndexNonExistant
	}
	m.databases[a], m.databases[b] = bSide, aSide
	m.save()
	return nil
}

func (m *manager) Get(id int) (Database, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if m.closed == true {
		return nil, ErrorManagerClosed
	}
	var db string
	var d Database
	db, ok := m.databases[id]
	if !ok {
		db = m.generate(id)
	}
	d, ok = m.pool[db]
	if !ok {
		d = NewDatabase(db)
		m.pool[db] = d
	}
	return d, nil
}

func (m *manager) Close() error {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.closed = true
	for _, v := range m.pool {
		err := v.Close()
		if err != nil {
			return err
		}
	}
	return m.DB.Close()
}
