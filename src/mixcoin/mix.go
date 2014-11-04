package mixcoin

import (
	"log"
	"time"
)

type Mix struct {
	sigc chan string
}

func NewMix() *Mix {
	mix := &Mix{
		sigc: make(chan string, 1),
	}
	mix.run()
	return mix
}

func (m *Mix) Put(msg *ChunkMessage) {
	delay := delay(msg.ReturnBy)
	go m.signal(delay, msg.OutAddr)
}

func (m *Mix) signal(delay int, addr string) {
	time.Sleep(time.Duration(delay*10) * time.Minute)
	m.sigc <- addr
}

func (m *Mix) run() {
	for {
		select {
		case addr := <-m.sigc:
			send(addr)
		}
	}
}

func delay(returnBy int) int {
	log.Printf("generating delay with returnby %d and currheight %d", returnBy, blockchainHeight)
	rand := randInt(returnBy - 1 - blockchainHeight)
	log.Printf("generated delay %v", rand)
	return 0
	//return currHeight + rand
}
