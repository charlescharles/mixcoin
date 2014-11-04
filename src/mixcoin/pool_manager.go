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

func (p *PoolManager) Get(t PoolType) (*PoolItem, error) {
	switch t {
	case Receivable:
		return nil, errors.New("cannot get poolitem from receivable pool")
	case Mixing:
		return p.mixing.Get()
	case Reserve:
		return p.reserve.Get()
	}
}

func (p *PoolManager) ReceivingKeys() []string {
	return p.receivable.Keys()
}

func (p *PoolManager) Scan(keys []string) []*PoolItem {
	return p.receivable.Scan(keys)
}

func (p *PoolManager) Filter(f func(*PoolItem) bool) {
	p.receivable.Filter(f)
}
