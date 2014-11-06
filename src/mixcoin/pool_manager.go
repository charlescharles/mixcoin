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

func (p *MixPoolManager) Put(t PoolType, item PoolItem) {
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
	switch t {
	case Receivable:
		return nil, errors.New("cannot get poolitem from receivable pool")
	case Mixing:
		return p.mixing.Get()
	case Reserve:
		return p.reserve.Get()
	default:
		return nil, errors.New("unhandled pooltype")
	}
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
