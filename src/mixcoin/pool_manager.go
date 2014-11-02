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
	bootstrapMixC  chan []*BoostrapMixChunk
	prune          chan bool

	mixingAddrs []string
)

func init() {
	pool = make(map[string]*Chunk)
	newChunkC = make(chan *NewChunk)
	receivedChunkC = make(chan *ReceivedChunk)
	requestChunkC = make(chan chan *Chunk)
	prune = make(chan bool)

	mixingAddrs = make([]string, 20)
}

func StartPoolManager() {
	log.Println("starting pool manager")

	go managePool()
	go signalPrune()
}

func managePool() {
	for {
		select {
		case newChunk := <-newChunkC:
			log.Printf("poolmgr handling new chunk: %v", newChunk)
			poolHandleNew(newChunk.addr, newChunk.chunkMsg)

		case receivedChunk := <-receivedChunkC:
			log.Printf("poolmgr handling received chunk: %v", receivedChunk)
			poolHandleReceived(receivedChunk.addr, receivedChunk.txInfo)

		case ch := <-requestChunkC:
			log.Printf("poolmgr handling chunk request: %v", ch)
			ch <- poolPopRandomMixingChunk()

		case <-prune:
			log.Printf("poolmgr pruning")
			poolPrune()

		case bootstrapChunks := <-bootstrapMixC:
			log.Printf("poolmgr bootstrapping with chunks %v", bootstrapChunks)
			poolHandleBootstrap(bootstrapChunks)
		}
	}
}

func poolHandleBootstrap(bootstrapChunks []*BootstrapMixChunk) {
	for _, bootstrapChunk := range bootstrapChunks {
		receivableForm := bootstrapChunk.toReceivable()
		chunk := &Chunk{
			status:  Mixing,
			message: nil,
			txInfo:  receivableForm.txInfo,
		}

		pool[receivableForm.addr] = chunk
		mixingAddrs = append(mixingAddrs, receivableForm.addr)
	}
}

func poolPrune() {
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
	log.Printf("picked random address %s at index %d", randAddr, randIndex)

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
	log.Printf("received chunk at txinfo: %v", txInfo)
	chunk, ok := pool[addr]
	if !ok {
		log.Printf("pool doesn't contain this address: %v", addr)
		return
	}
	chunk.txInfo = txInfo
	log.Printf("assigned txinfo")

	pool[addr].status = Mixing
	mixingAddrs = append(mixingAddrs, addr)
	log.Printf("added address %s to mixing pool", addr)
	randDelay := generateDelay(chunk.message.ReturnBy)
	log.Printf("generated delay: %v blocks", randDelay)
	outAddr := chunk.message.OutAddr

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
