package mixcoin

type ReceivingPool struct {
	putc  chan PoolItem
	recvc chan []string
	scanc chan []*PoolItem
}

func NewReceivingPool() *ReceivingPool {
	p := &ReceivingPool{
		putc:  make(chan PoolItem),
		recvc: make(chan []string),
		scanc: make(chan []*PoolItem),
	}
	go p.run()
	return p
}

func (p *ReceivingPool) Put(item PoolItem) {
	p.putc <- item
}

func (p *ReceivingPool) Scan(keys []string) []*PoolItem {
	p.recvc <- keys
	return <-p.scanc
}

func (p *ReceivingPool) Keys() []string {
	ch := make(chan []string)
	p.keysc <- ch
	return <-ch
}

func (p *ReceivingPool) run() {
	table := make(map[string]*PoolItem)

	for {
		select {
		case item := <-p.putc:
			table[item.Key()] = item

		case ch := <-p.keysc:
			var keys []string
			for _, key := range keys {
				keys = append(keys, key)
			}
			ch <- keys

		case keys := <-p.recvc:
			var items []*PoolItem
			for _, key := range keys {
				item, ok := table[key]
				if ok {
					items = append(items, item)
					delete(table, key)
				}
			}
			p.scanc <- items
		}
	}
}
