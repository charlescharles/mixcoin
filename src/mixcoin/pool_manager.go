package mixcoin

// TODO use crypto/rand
import (
	"btcutil"
	"errors"
	"math/rand"
	"sync"
)

type PoolType int

const (
	Receivable PoolType = iota
	Mixing
	Retained
)

type NewChunk struct {
	addr  string
	chunk *Chunk
}

type ReceivedChunk struct {
	addr   string
	txInfo *TxInfo
}

var (
	pool           map[string]*Chunk
	newChunkC      chan *NewChunk
	receivedChunkC chan *ReceivedChunk
	mixingAddrs    []string
)

func StartPoolManager() {
	pool = make(map[string]*Chunk)
	newChunkC = make(chan *NewChunk)
	receivedChunkC = make(chan *ReceivedChunk)
	mixingAddrs = make([]string)

	go managePool()
}

func managePool() {
	for {
		select {
		case newChunk := <-newChunkC:
			ch := Chunk{}
			ch.ChunkMessage = newChunk.chunk
			pool[newChunk.addr] = ch
		case receivedChunk := <-receivedChunkC:
			pool[receivedChunk.addr].txInfo = receivedChunk.txInfo
			pool[receivedChunk.addr].status = Mixing
			mixingAddrs = append(mixingAddrs, receivedChunk.addr)
		}
	}
}
