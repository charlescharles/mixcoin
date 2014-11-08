package mixcoin

import (
	"errors"
)

type PoolItem interface {
	Key() string
	Serialize() []byte
}

type RandomizingPool struct {
	putc chan PoolItem
	getc chan chan PoolItem
}

func NewRandomizingPool() *RandomizingPool {
	p := &RandomizingPool{
		putc: make(chan PoolItem),
		getc: make(chan chan PoolItem),
	}

	go p.run()

	return p
}

func (p *RandomizingPool) Put(item PoolItem) {
	p.putc <- item
}

func (p *RandomizingPool) Get() (PoolItem, error) {
	ch := make(chan PoolItem)
	p.getc <- ch
	ret := <-ch
	if ret == nil {
		return nil, errors.New("pool is empty")
	}
	return ret, nil
}

func (p *RandomizingPool) run() {
	var keys []string
	table := make(map[string]PoolItem)

	for {
		select {
		case item := <-p.putc:
			keys = append(keys, item.Key())
			table[item.Key()] = item

		case ch := <-p.getc:
			n := len(keys)
			if n == 0 {
				ch <- nil
				continue
			}
			rand := randInt(n)
			key := keys[rand]

			ret := table[key]
			delete(table, key)
			keys[rand] = keys[n-1]
			keys = keys[:n-1]

			ch <- ret
		}
	}
}
