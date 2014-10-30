package mixcoin

import (
	"log"
	"time"
)

// how often to prune expired receivable chunks, in seconds
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
	pool                map[string]*Chunk
	newChunkC           chan *NewChunk
	receivedChunkC      chan *ReceivedChunk
	requestMixingChunkC chan bool
	randMixingChunkC    chan *Chunk
	prune               chan bool

	mixingAddrs []string
)

func StartPoolManager() {
	log.Println("starting pool manager")

	pool = make(map[string]*Chunk)
	newChunkC = make(chan *NewChunk)
	receivedChunkC = make(chan *ReceivedChunk)
	requestMixingChunkC = make(chan bool)
	randMixingChunkC = make(chan *Chunk)
	prune = make(chan bool)

	mixingAddrs = make([]string, 10)

	go managePool()
	go signalPrune()
}

func managePool() {
	for {
		select {
		case newChunk := <-newChunkC:
			log.Printf("adding new chunk: %v", newChunk)
			ch := &Chunk{Receivable, newChunk.chunkMsg, nil}
			pool[newChunk.addr] = ch
		case receivedChunk := <-receivedChunkC:
			// change chunk to mixing, add txoutinfo, add chunk to mixingaddrs,
			// mix
			log.Printf("received chunk: %v", receivedChunk)
			pool[receivedChunk.addr].txInfo = receivedChunk.txInfo
			pool[receivedChunk.addr].status = Mixing
			mixingAddrs = append(mixingAddrs, receivedChunk.addr)
			randDelay := generateDelay(pool[receivedChunk.addr].message.ReturnBy)
			outAddr := receivedChunk.addr
			go mix(randDelay, outAddr)
		case <-requestMixingChunkC:
			randIndex := randInt(len(mixingAddrs))
			randAddr := mixingAddrs[randIndex]

			chunk := pool[randAddr]
			delete(pool, randAddr)

			// remove from mixingAddrs
			mixingAddrs[randIndex] = mixingAddrs[len(mixingAddrs)-1]
			mixingAddrs = mixingAddrs[:len(mixingAddrs)-1]

			randMixingChunkC <- chunk
		case <-prune:
			expiredAddrs := make([]string, 10)
			for addr, chunk := range pool {
				if chunk.status == Receivable && isExpired(chunk) {
					expiredAddrs = append(expiredAddrs, addr)
				}
			}
			for _, addr := range expiredAddrs {
				delete(pool, addr)
			}
		}
	}
}

func isExpired(chunk *Chunk) bool {
	currHeight, _ := getBlockchainHeight()
	isPastExpiry := chunk.message.SendBy < currHeight

	return isPastExpiry
}

func signalPrune() {
	for {
		time.Sleep(time.Duration(PRUNE_PERIOD) * time.Second)
		prune <- true
	}
}
