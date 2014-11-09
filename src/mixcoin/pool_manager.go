package mixcoin

import (
	"errors"
)

type PoolType int

const (
	Receivable PoolType = iota
	Mixing
	Reserve
)

type PoolManager interface {
	Put(PoolType, PoolItem)
	Get(PoolType) (PoolItem, error)
	ReceivingKeys() []string
	Scan([]string) []PoolItem
	Filter(func(PoolItem) bool)
	Shutdown()
}

type MixPoolManager struct {
	receivable *ReceivingPool
	mixing     *RandomizingPool
	reserve    *RandomizingPool
}

func NewPoolManager() PoolManager {
	return &MixPoolManager{
		receivable: NewReceivingPool(),
		mixing:     NewRandomizingPool(),
		reserve:    NewRandomizingPool(),
	}
}

func (p *MixPoolManager) Shutdown() {
	// not sure what to do here -- should this do anything?
}

func (p *MixPoolManager) Put(t PoolType, item PoolItem) {
	db.Put(item)
	switch t {
	case Receivable:
		p.receivable.Put(item)
	case Mixing:
		p.mixing.Put(item)
	case Reserve:
		p.reserve.Put(item)
	}
}

func (p *MixPoolManager) Get(t PoolType) (PoolItem, error) {
	var item PoolItem
	item = nil
	var err error
	switch t {
	case Receivable:
		err = errors.New("cannot get poolitem from receivable pool")
	case Mixing:
		item, err = p.mixing.Get()
	case Reserve:
		item, err = p.reserve.Get()
	default:
		err = errors.New("unhandled pooltype")
	}

	if err != nil {
		db.Delete(item.Key())
	}

	return item, err
}

func (p *MixPoolManager) ReceivingKeys() []string {
	return p.receivable.Keys()
}

func (p *MixPoolManager) Scan(keys []string) []PoolItem {
	return p.receivable.Scan(keys)
}

func (p *MixPoolManager) Filter(f func(PoolItem) bool) {
	p.receivable.Filter(f)
}
