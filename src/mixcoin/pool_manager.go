package mixcoin

import (
	"log"
	"time"
)

// how often to prune expired receivable chunks, in minutes
const PRUNE_PERIOD = 10

type PoolType int

const (
	Receivable PoolType = iota
	Mixing
	Retained
)

type NewChunk struct {
	addr     string
	chunkMsg *ChunkMessage
}

type ReceivedChunk struct {
	addr   string
	txInfo *TxInfo
}

var (
	pool           map[string]*Chunk
	newChunkC      chan *NewChunk
	receivedChunkC chan *ReceivedChunk
	requestChunkC  chan chan *Chunk
	prune          chan bool

	mixingAddrs []string
)

func StartPoolManager() {
	log.Println("starting pool manager")

	go managePool()
	go signalPrune()
}

func managePool() {
	for {
		select {
		case newChunk := <-newChunkC:
			poolHandleNew(newChunk.addr, newChunk.chunkMsg)

		case receivedChunk := <-receivedChunkC:
			poolHandleReceived(receivedChunk.addr, receivedChunk.txInfo)

		case ch := <-requestMixingChunkC:
			ch <- poolPopRandomMixingChunk()

		case <-prune:
			poolPrune()
		}
	}
}

func poolPrune() {
	log.Printf("pruning expired receivable chunks")

	expiredAddrs := make([]string, 10)
	for addr, chunk := range pool {
		if chunk.status == Receivable && isExpired(chunk) {
			log.Printf("found expired chunk: %s", addr)
			expiredAddrs = append(expiredAddrs, addr)
		}
	}
	for _, addr := range expiredAddrs {
		delete(pool, addr)
	}
}

func poolHandleNew(addr string, chunkMsg *ChunkMessage) {
	chunk := &Chunk{Receivable, chunkMsg, nil}
	pool[addr] = chunk
}

func poolPopRandomMixingChunk() *Chunk {
	randIndex := randInt(len(mixingAddrs))
	randAddr := mixingAddrs[randIndex]

	chunk := pool[randAddr]
	log.Printf("popping rand chunk: %v", chunk)
	delete(pool, randAddr)
	log.Printf("the chunk is still %v", chunk)

	// remove from mixingAddrs
	mixingAddrs[randIndex] = mixingAddrs[len(mixingAddrs)-1]
	mixingAddrs = mixingAddrs[:len(mixingAddrs)-1]
	return chunk
}

func poolHandleReceived(addr string, txInfo *TxInfo) {
	// change chunk to mixing, add txoutinfo, add chunk to mixingaddrs,
	// mix
	log.Printf("received chunk: %v", receivedChunk)
	log.Printf("with txinfo: %v", receivedChunk.txInfo)
	chunk, ok := pool[receivedChunk.addr]
	if !ok {
		log.Printf("pool doesn't contain this address: %v", receivedChunk.addr)
		return
	}
	chunk.txInfo = receivedChunk.txInfo
	log.Printf("assigned txinfo")

	pool[receivedChunk.addr].status = Mixing
	mixingAddrs = append(mixingAddrs, receivedChunk.addr)
	log.Printf("added address %s to mixing pool", receivedChunk.addr)
	randDelay := generateDelay(pool[receivedChunk.addr].message.ReturnBy)
	log.Printf("generated delay: %v blocks", randDelay)
	outAddr := receivedChunk.addr

	go mix(randDelay, outAddr)
}

func isExpired(chunk *Chunk) bool {
	currHeight, _ := getBlockchainHeight()
	isPastExpiry := chunk.message.SendBy <= currHeight

	return isPastExpiry
}

func signalPrune() {
	for {
		time.Sleep(time.Duration(PRUNE_PERIOD) * time.Minute)
		prune <- true
	}
}
