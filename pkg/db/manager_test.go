package db_test

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/furui/gochunk/pkg/db"
	"github.com/furui/gochunk/pkg/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/furui/gochunk/pkg/config"
)

func init() {
	files, err := filepath.Glob(filepath.Join(os.TempDir(), "*.db"))
	if err != nil {
		return
	}
	for _, f := range files {
		err = os.Remove(f)
		if err != nil {
			log.Printf("warning: couldn't remove %s: %s", f, err)
		}
	}
}

func TestGet(t *testing.T) {
	conf := config.NewConfig()
	conf.DatabaseLocation = os.TempDir()
	manager := db.NewManager(conf, uuid.NewGenerator())
	assert.NotPanics(t, func() {
		first, err := manager.Get(1)
		assert.NoError(t, err)
		assert.NotNil(t, first)
		second, err := manager.Get(5)
		assert.NoError(t, err)
		assert.NotNil(t, second)
		third, err := manager.Get(1)
		assert.NoError(t, err)
		assert.Equal(t, first, third)
	})
	err := manager.Swap(3, 4)
	assert.Error(t, err)
	err = manager.Swap(1, 2)
	assert.Error(t, err)
	err = manager.Swap(5, 2)
	assert.Error(t, err)
	err = manager.Swap(1, 5)
	assert.NoError(t, err)
	err = manager.Close()
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(os.TempDir(), "manager.db"))

	manager = db.NewManager(conf, uuid.NewGenerator())
	assert.NotPanics(t, func() {
		first, err := manager.Get(1)
		assert.NoError(t, err)
		assert.NotNil(t, first)
	})

}
