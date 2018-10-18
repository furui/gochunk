package db

import (
	"bytes"
	"errors"

	bbolt "github.com/etcd-io/bbolt"
	respTypes "github.com/furui/gochunk/pkg/types"
)

type Encoding int

const (
	EncodeRaw Encoding = iota
	EncodeInt
)

var (
	// ErrKeyError is thrown when the key cannot be returned
	ErrKeyError = errors.New("internal key error")
)

type Key struct {
	Name       []byte
	Type       Encoding
	Data       []byte
	Expiration int64
}

func (d *database) Key(name []byte) (*Key, error) {
	var data []byte
	err := d.DB.View(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("keys"))
		if err != nil {
			return err
		}
		data = b.Get(name)
		return nil
	})
	scanner := respTypes.NewScanner(bytes.NewBuffer(data))
	res := scanner.Scan()
	if res != true {
		return nil, ErrKeyError
	}
	val, ok := scanner.Type().(*respTypes.Array)
	if !ok {
		return nil, ErrKeyError
	}
	contents := val.Contents
	if len(contents) < 4 {
		return nil, ErrKeyError
	}
	nam, ok := contents[0].Value().([]byte)
	if !ok {
		return nil, ErrKeyError
	}
	enc, ok := contents[1].Value().(Encoding)
	if !ok {
		return nil, ErrKeyError
	}
	dat, ok := contents[2].Value().([]byte)
	if !ok {
		return nil, ErrKeyError
	}
	exp, ok := contents[3].Value().(int64)
	return &Key{
		Name:       nam,
		Type:       enc,
		Data:       dat,
		Expiration: exp,
	}, err
}
