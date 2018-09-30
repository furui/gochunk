package uuid_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/furui/gochunk/pkg/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIterations(t *testing.T) {
	generator := uuid.NewGenerator()
	uuids := make(map[string]int, 0)
	for i := 0; i < 10; i++ {
		val := generator.GenerateTimeCounter()
		if _, ok := uuids[val.String()]; ok {
			t.Errorf("Non-unique value %s", val)
		} else {
			parts := strings.Split(val.String(), "-")
			assert.Len(t, parts[0], 8)
			assert.Len(t, parts[1], 4)
			assert.Len(t, parts[2], 4)
			assert.Len(t, parts[3], 4)
			assert.Len(t, parts[4], 12)
			uuids[val.String()] = 1
		}
	}
}

func TestCounter(t *testing.T) {
	generator := uuid.NewGenerator()
	first := generator.GenerateTimeCounter().String()
	generator.Reset()
	second := generator.GenerateTimeCounter().String()
	third := generator.GenerateTimeCounter().String()
	lastDigit := first[len(first)-1:]
	assert.Equal(t, lastDigit, "0")
	lastDigit = second[len(second)-1:]
	assert.Equal(t, lastDigit, "0")
	lastDigit = third[len(third)-1:]
	assert.Equal(t, lastDigit, "1")
}

func TestVersion(t *testing.T) {
	generator := uuid.NewGenerator()
	uuid := generator.GenerateTimeCounter()
	assert.Equal(t, uuid.Version(), 2)
}

func TestVariant(t *testing.T) {
	generator := uuid.NewGenerator()
	uuid := generator.GenerateTimeCounter()
	assert.Equal(t, uuid.Variant(), 0)
}

func TestConcurrent(t *testing.T) {
	generator := uuid.NewGenerator()
	uuidChan := make(chan uuid.UUID, 1)
	uuids := make(map[string]int, 0)
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(gen uuid.Generator, ch chan uuid.UUID, wg *sync.WaitGroup) {
			defer wg.Done()
			ch <- gen.GenerateTimeCounter()
		}(generator, uuidChan, &wg)
	}
	go func(t *testing.T, ch chan uuid.UUID, uuids *map[string]int) {
		for u := range ch {
			if _, ok := (*uuids)[u.String()]; ok {
				t.Errorf("Non-unique value %s", u.String())
			} else {
				(*uuids)[u.String()] = 1
			}
		}
	}(t, uuidChan, &uuids)
	wg.Wait()
}
