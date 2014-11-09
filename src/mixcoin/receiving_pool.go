package mixcoin

type ReceivingPool struct {
	putc    chan PoolItem
	recvc   chan []string
	scanc   chan []PoolItem
	keysc   chan chan []string
	filterc chan func(PoolItem) bool
}

func NewReceivingPool() *ReceivingPool {
	p := &ReceivingPool{
		putc:    make(chan PoolItem),
		recvc:   make(chan []string),
		scanc:   make(chan []PoolItem),
		keysc:   make(chan chan []string),
		filterc: make(chan func(PoolItem) bool),
	}
	go p.run()
	return p
}

func (p *ReceivingPool) Put(item PoolItem) {
	p.putc <- item
}

func (p *ReceivingPool) Scan(keys []string) []PoolItem {
	p.recvc <- keys
	return <-p.scanc
}

func (p *ReceivingPool) Keys() []string {
	ch := make(chan []string)
	p.keysc <- ch
	return <-ch
}

func (p *ReceivingPool) Filter(f func(PoolItem) bool) {
	p.filterc <- f
}

func (p *ReceivingPool) run() {
	table := make(map[string]PoolItem)

	for {
		select {
		case item := <-p.putc:
			table[item.Key()] = item

		case ch := <-p.keysc:
			var keys []string
			for key, _ := range table {
				keys = append(keys, key)
			}
			ch <- keys

		case keys := <-p.recvc:
			var items []PoolItem
			for _, key := range keys {
				item, ok := table[key]
				if ok {
					items = append(items, item)
					delete(table, key)
				}
			}
			p.scanc <- items

		case f := <-p.filterc:
			var remove []string
			for key, item := range table {
				if f(item) {
					remove = append(remove, key)
				}
			}
			for _, key := range remove {
				delete(table, key)
				db.Delete(key)
			}
		}
	}
}
