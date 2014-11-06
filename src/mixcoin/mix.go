package mixcoin

import (
	"log"
	"time"
)

type Mix struct {
	debugc chan string
	debug  bool
}

func NewMix(debugc chan string) *Mix {
	mix := &Mix{}
	if debugc != nil {
		mix.debugc = debugc
		mix.debug = true
	}
	return mix
}

func (m *Mix) Put(msg *ChunkMessage) {
	delay := generateDelay(msg.ReturnBy)
	go m.signal(delay, msg.OutAddr)
}

func (m *Mix) signal(delay int, addr string) {
	time.Sleep(time.Duration(delay*10) * time.Minute)
	if m.debug {
		m.debugc <- addr
	} else {
		go send(addr)
	}
}

func generateDelay(returnBy int) int {
	log.Printf("generating delay with returnby %d and currheight %d", returnBy, blockchainHeight)
	rand := randInt(returnBy - 1 - blockchainHeight)
	log.Printf("generated delay %v", rand)
	return 0
	//return currHeight + rand
}
