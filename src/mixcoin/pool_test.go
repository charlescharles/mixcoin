package mixcoin

import (
	"github.com/stretchr/testify/assert"
	"log"
	"sort"
	"testing"
)

type Pair struct {
	key string
	val string
}

func (m *Pair) Key() string {
	return m.key
}

var (
	items = []*Pair{
		&Pair{"gorilla", "primate"},
		&Pair{"giraffe", "long-necked animal"},
		&Pair{"bat", "mammal"},
		&Pair{"jellyfish", "cnidarian"},
		&Pair{"shark", "fish"},
	}
)

func TestRandomizingPool(t *testing.T) {
	p := NewRandomizingPool()
	keys := make(map[string]int)
	for _, item := range items {
		keys[item.Key()] = 0
		p.Put(item)
	}

	for i := 0; i < len(items); i++ {
		item, err := p.Get()
		assert.Nil(t, err, "there should be items remaining")
		count, ok := keys[item.Key()]
		assert.True(t, ok, "key should be present")
		assert.Equal(t, count, 0, "key shouldn't have been seen yet")

		keys[item.Key()] += 1
	}

	_, err := p.Get()
	assert.NotNil(t, err, "pool should empty now")
}

func TestReceivingPool(t *testing.T) {
	p := NewReceivingPool()

	var keys []string
	for _, item := range items {
		p.Put(item)
		keys = append(keys, item.Key())
	}
	gotKeys := p.Keys()
	log.Printf("got keys: %v", gotKeys)
	log.Printf("want keys: %v", keys)
	assert.True(t, setEquals(keys, gotKeys), "pool.Keys should return correct slice of keys")

	gotKeys = nil
	subset := []string{"giraffe", "jellyfish", "shark"}
	for _, item := range p.Scan(subset) {
		gotKeys = append(gotKeys, item.Key())
	}
	log.Printf("got keys: %v", gotKeys)
	log.Printf("want keys: %v", subset)
	assert.True(t, setEquals(subset, gotKeys), "pool.Scan should return correct intersection")
}

func setEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	sort.Strings(a)
	sort.Strings(b)

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
