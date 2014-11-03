package mixcoin

import (
	"btcutil"
	"btcwire"
	"errors"
	"log"
	"math/big"
	"math/rand"
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

type ReceivedChunk struct {
	addr      string
	txInfo    *TxInfo
	blockHash *btcwire.ShaHash
}

var (
	pool                map[string]*Chunk
	newChunkC           chan *ChunkMessage
	receivedChunkC      chan *ReceivedChunk
	requestMixChunkC    chan chan *Chunk
	requestReceivablesC chan chan []btcutil.Address
	requestFeeChunkC    chan chan *Chunk
	addFeeChunkC        chan *Chunk
	bootstrapFeeC       chan []*BootstrapFeeChunk
	prune               chan bool

	mixingAddrs   []string
	retainedAddrs []string
)

func init() {
	pool = make(map[string]*Chunk)
	newChunkC = make(chan *ChunkMessage)
	receivedChunkC = make(chan *ReceivedChunk)
	requestMixChunkC = make(chan chan *Chunk)
	requestReceivablesC = make(chan chan []btcutil.Address)
	requestFeeChunkC = make(chan chan *Chunk)
	addFeeChunkC = make(chan *Chunk)
	prune = make(chan bool)

	mixingAddrs = make([]string, 20)
	retainedAddrs = make([]string, 20)
}

func StartPoolManager() {
	log.Println("starting pool manager")

	go managePool()
	go signalPrune()
}

func managePool() {
	for {
		select {
		case newChunkMsg := <-newChunkC:
			log.Printf("poolmgr handling new chunk: %v", newChunkMsg)
			poolHandleNew(newChunkMsg)

		case receivedChunk := <-receivedChunkC:
			log.Printf("poolmgr handling received chunk: %v", receivedChunk)
			poolHandleReceived(receivedChunk.addr, receivedChunk.txInfo, receivedChunk.blockHash)

		case ch := <-requestMixChunkC:
			log.Printf("poolmgr handling chunk request: %v", ch)
			ch <- poolPopRandomMixingChunk()

		case ch := <-requestReceivablesC:
			log.Printf("poolmgr handling request for receivable chunks")
			ch <- poolGetReceivableChunks()

		case ch := <-requestFeeChunkC:
			log.Printf("poolmgr handling request for fee chunk")
			ch <- poolPopRandomMixingChunk()

		case newFeeChunk := <-addFeeChunkC:
			log.Printf("poolmgr adding fee chunk")
			poolAddFeeChunk(newFeeChunk)

		case <-prune:
			log.Printf("poolmgr pruning")
			poolPrune()

		case bootstrapChunks := <-bootstrapFeeC:
			log.Printf("poolmgr bootstrapping with chunks %v", bootstrapChunks)
			poolHandleBootstrap(bootstrapChunks)
		}
	}
}

func addChunkToPool(chunkMsg *ChunkMessage) {
	newChunkC <- chunkMsg
}

func getReceivableChunks() []btcutil.Address {
	ch := make(chan []btcutil.Address)
	requestReceivablesC <- ch
	receivableAddrs := <-ch
	return receivableAddrs
}

func markReceived(addr string, txInfo *TxInfo, blockHash *btcwire.ShaHash) {
	receivedChunkC <- &ReceivedChunk{addr, txInfo, blockHash}
}

func getMixChunk() *Chunk {
	ch := make(chan *Chunk)
	requestMixChunkC <- ch
	output := <-ch
	return output
}

func getFeeChunk() *Chunk {
	ch := make(chan *Chunk)
	requestFeeChunkC <- ch
	output := <-ch
	return output
}

func bootstrapFeeChunks(chunks []*BootstrapFeeChunk) {
	bootstrapFeeC <- chunks
}

func poolAddFeeChunk(chunk *Chunk) {
	chunk.status = Retained
	pool[chunk.addr] = chunk
	retainedAddrs = append(retainedAddrs, chunk.addr)
}

func poolGetReceivableChunks() []btcutil.Address {
	receivableAddrs := make([]btcutil.Address, 10)
	for addr, chunk := range pool {
		if chunk.status == Receivable {
			decoded, err := decodeAddress(addr)
			if err != nil {
				log.Panicf("unable to decode address: %v", err)
			}
			receivableAddrs = append(receivableAddrs, decoded)
		}
	}
	return receivableAddrs
}

func poolHandleBootstrap(bootstrapChunks []*BootstrapFeeChunk) {
	for _, bootstrapChunk := range bootstrapChunks {
		chunk, wif, err := bootstrapChunk.normalize()

		if err != nil {
			log.Printf("error parsing bootstrap chunk: %v", err)
		}

		err = getRpcClient().ImportPrivKey(wif)

		if err != nil {
			log.Printf("error importing privkey: %v", err)
		}

		pool[chunk.addr] = chunk
		retainedAddrs = append(retainedAddrs, chunk.addr)
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

func poolHandleNew(chunkMsg *ChunkMessage) {
	addr := chunkMsg.MixAddr
	chunk := &Chunk{
		status:  Receivable,
		message: chunkMsg,
		txInfo:  nil,
		addr:    addr,
	}
	pool[addr] = chunk
}

func poolPopRandomMixingChunk() *Chunk {
	log.Printf("generating random index in [0, %d)", len(mixingAddrs))
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

func poolHandleReceived(addr string, txInfo *TxInfo, blockHash *btcwire.ShaHash) {
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

	if isFee(chunk.message.Nonce, blockHash, chunk.message.Fee) {
		chunk.status = Retained
		retainedAddrs = append(retainedAddrs, addr)
		return
	}

	chunk.status = Mixing
	mixingAddrs = append(mixingAddrs, addr)
	log.Printf("added address %s to mixing pool", addr)
	randDelay := generateDelay(chunk.message.ReturnBy)
	log.Printf("generated delay: %v blocks", randDelay)
	outAddr := chunk.message.OutAddr

	go mix(randDelay, outAddr)
}

func poolPopRandomFeeChunk() (*Chunk, error) {
	log.Printf("generating random index in [0, %d)", len(retainedAddrs))
	if len(retainedAddrs) == 0 {
		return nil, errors.New("pool has no retained chunks")
	}
	randIndex := randInt(len(retainedAddrs))
	randAddr := retainedAddrs[randIndex]
	log.Printf("picked random address %s at index %d", randAddr, randIndex)

	chunk := pool[randAddr]
	log.Printf("popping rand chunk: %v", chunk)
	delete(pool, randAddr)
	log.Printf("the chunk is still %v", chunk)

	// remove from retainedAddrs
	retainedAddrs[randIndex] = retainedAddrs[len(retainedAddrs)-1]
	retainedAddrs = retainedAddrs[:len(retainedAddrs)-1]
	return chunk, nil
}

func isFee(nonce int64, hash *btcwire.ShaHash, feeBips int) bool {
	bigIntHash := big.NewInt(0)
	bigIntHash.SetBytes(hash.Bytes())
	hashInt := bigIntHash.Int64()

	gen := nonce | hashInt
	fee := float64(feeBips) * 1.0e-4

	source := rand.NewSource(gen)
	rng := rand.New(source)
	return rng.Float64() <= fee
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
