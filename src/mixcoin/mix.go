package mixcoin

import (
	"log"
	"time"
)

type Mix struct {
	debugc    chan string
	debug     bool
	shutdownc []chan struct{}
}

func NewMix(debugc chan string) *Mix {
	mix := &Mix{}
	if debugc != nil {
		mix.debugc = debugc
		mix.debug = true
	}
	return mix
}

func (m *Mix) Shutdown() {
	log.Printf("shutting down mix...")
	for _, c := range m.shutdownc {
		c <- struct{}{}
	}
}

func (m *Mix) Put(msg *ChunkMessage) {
	delay := generateDelay(msg.ReturnBy)

	// TODO: pretty sure these should all be unbuffered chans
	// because we want Mix#Shutdown to block
	shutdown := make(chan struct{})
	m.shutdownc = append(m.shutdownc, shutdown)
	go m.signal(delay, msg.OutAddr, shutdown)
}

func (m *Mix) signal(delay int, addr string, shutdown chan struct{}) {
	if m.debug {
		m.debugc <- addr
	} else {
		select {
		case <-time.After(time.Duration(delay*10) * time.Minute):
			go send(addr)
		case <-shutdown:
			log.Printf("sending chunk early as part of shutdown")
			go send(addr)
		}
	}
}

func generateDelay(returnBy int) int {
	log.Printf("generating delay with returnby %d and currheight %d", returnBy, blockchainHeight)
	rand := randInt(returnBy - 1 - blockchainHeight)
	log.Printf("generated delay %v", rand)
	return 0
	//return currHeight + rand
}
