package mixcoin

// TODO use crypto/rand
import (
	"log"
	"math/rand"
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
			log.Println("adding new chunk")
			ch := &Chunk{Receivable, newChunk.chunkMsg, nil}
			pool[newChunk.addr] = ch
		case receivedChunk := <-receivedChunkC:
			pool[receivedChunk.addr].txInfo = receivedChunk.txInfo
			pool[receivedChunk.addr].status = Mixing
			mixingAddrs = append(mixingAddrs, receivedChunk.addr)
			randDelay := generateDelay(pool[receivedChunk.addr].message.ReturnBy)
			outAddr := receivedChunk.addr
			go mix(randDelay, outAddr)
		case <-requestMixingChunkC:
			// TODO remove randAddr from mixingAddrs
			randIndex := rand.Intn(len(mixingAddrs))
			randAddr := mixingAddrs[randIndex]
			chunk := pool[randAddr]
			delete(pool, randAddr)
			randMixingChunkC <- chunk
		case <-prune:
			expiredAddrs := make([]string, 10)
			for addr, chunk := range pool {
				if isExpired(chunk) {
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
	isReceivable := chunk.status == Receivable
	isPastExpiry := false
	return isReceivable && isPastExpiry
}

func signalPrune() {
	for {
		time.Sleep(PRUNE_PERIOD * time.Second)
		prune <- true
	}
}
