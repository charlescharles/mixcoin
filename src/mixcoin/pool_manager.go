package main

type PoolType int

const (
	Receivable PoolType = iota
	Mixing
	Reserve
)

type PoolManager struct {
	receivable *ReceivingPool
	mixing     *RandomizingPool
	reserve    *RandomizingPool
}

func NewPoolManager() *PoolManager {
	return &PoolManager{
		receivable: NewReceivingPool(),
		mixing:     NewRandomizingPool(),
		reserve:    NewRandomizingPool(),
	}
}

func (p *PoolManager) Put(t PoolType, item *PoolItem) {
	switch t {
	case Receivable:
		p.receivable.Put(item)
	case Mixing:
		p.mixing.Put(item)
	case Reserve:
		p.reserve.Put(item)
	}
}
