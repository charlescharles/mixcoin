package mixcoin

import (
	"errors"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"log"
	"math/big"
	"math/rand"
)

var pool PoolManager

func GetPool() PoolManager {
	return pool
}

func StartPoolManager() {
	log.Println("starting pool manager")
	pool = NewPoolManager()
}

type ReceivedChunk struct {
	addr      string
	txInfo    *TxInfo
	blockHash *btcwire.ShaHash
}

type PoolType int

const (
	Receivable PoolType = iota
	Mixing
	Reserve
)

type PoolManagerDaemon struct {
	pool                 map[string]*Chunk
	newChunkC            chan *ChunkMessage
	receivedChunkC       chan *ReceivedChunk
	requestMixChunkC     chan chan *Chunk
	requestReceivablesC  chan chan []btcutil.Address
	requestReserveChunkC chan chan *Chunk
	addFeeChunkC         chan *Chunk
	bootstrapFeeC        chan []*BootstrapFeeChunk
	pruneSig             chan int

	mixingAddrs   []string
	retainedAddrs []string
}

func NewPoolManager() PoolManager {
	// don't initialize mixingAddrs and retainedAddrs because they're zerolength slices
	poolMgr := &PoolManagerDaemon{
		pool:                 make(map[string]*Chunk),
		newChunkC:            make(chan *ChunkMessage),
		receivedChunkC:       make(chan *ReceivedChunk),
		requestMixChunkC:     make(chan chan *Chunk),
		requestReceivablesC:  make(chan chan []btcutil.Address),
		requestReserveChunkC: make(chan chan *Chunk),
		addFeeChunkC:         make(chan *Chunk),
		bootstrapFeeC:        make(chan []*BootstrapFeeChunk),
		pruneSig:             make(chan int),
	}

	go poolMgr.manage()

	return poolMgr
}

func (poolMgr *PoolManagerDaemon) manage() {
	for {
		select {
		case newChunkMsg := <-poolMgr.newChunkC:
			log.Printf("poolmgr handling new chunk: %v", newChunkMsg)
			poolMgr.handleNew(newChunkMsg)

		case receivedChunk := <-poolMgr.receivedChunkC:
			log.Printf("poolmgr handling received chunk: %v", receivedChunk)
			poolMgr.handleReceived(receivedChunk.addr, receivedChunk.txInfo, receivedChunk.blockHash)

		case ch := <-poolMgr.requestMixChunkC:
			log.Printf("poolmgr handling chunk request: %v", ch)
			ch <- poolMgr.popRandomMixingChunk()

		case ch := <-poolMgr.requestReceivablesC:
			log.Printf("poolmgr handling request for receivable chunks")
			ch <- poolMgr.getReceivableChunks()

		case ch := <-poolMgr.requestReserveChunkC:
			log.Printf("poolmgr handling request for fee chunk")
			ret, err := poolMgr.popRandomReserveChunk()
			if err != nil {
				log.Printf("error popping rand reserve chunk: %v", err)
			}
			ch <- ret

		case newFeeChunk := <-poolMgr.addFeeChunkC:
			log.Printf("poolmgr adding fee chunk")
			poolMgr.addFeeChunk(newFeeChunk)

		case height := <-poolMgr.pruneSig:
			log.Printf("poolmgr pruning")
			poolMgr.prune(height)

		case chunks := <-poolMgr.bootstrapFeeC:
			log.Printf("poolmgr bootstrapping chunks")
			poolMgr.handleBootstrap(chunks)
		}
	}
}

func (poolMgr *PoolManagerDaemon) BootstrapReserve(chunks []*BootstrapFeeChunk) {
	poolMgr.bootstrapFeeC <- chunks
}

func (poolMgr *PoolManagerDaemon) RegisterReserveChunk(chunk *Chunk) {
	poolMgr.addFeeChunkC <- chunk
}

func (poolMgr *PoolManagerDaemon) RegisterNewChunk(chunkMsg *ChunkMessage) {
	poolMgr.newChunkC <- chunkMsg
}

func (poolMgr *PoolManagerDaemon) GetReceivable() []btcutil.Address {
	ch := make(chan []btcutil.Address)
	poolMgr.requestReceivablesC <- ch
	receivableAddrs := <-ch
	return receivableAddrs
}

func (poolMgr *PoolManagerDaemon) RegisterReceived(addr string, txInfo *TxInfo, blockHash *btcwire.ShaHash) {
	poolMgr.receivedChunkC <- &ReceivedChunk{addr, txInfo, blockHash}
}

func (poolMgr *PoolManagerDaemon) GetRandomChunk(poolType PoolType) (*Chunk, error) {
	ch := make(chan *Chunk)
	switch poolType {
	case Mixing:
		poolMgr.requestMixChunkC <- ch
	case Reserve:
		poolMgr.requestReserveChunkC <- ch
	default:
		return nil, errors.New("unhandled pooltype")
	}
	output := <-ch
	return output, nil
}

func (poolMgr *PoolManagerDaemon) Prune(height int) {
	poolMgr.pruneSig <- height
}

func (poolMgr *PoolManagerDaemon) bootstrapFeeChunks(chunks []*BootstrapFeeChunk) {
	log.Printf("sending bootstrap chunks to poolmgr")
	poolMgr.bootstrapFeeC <- chunks
}

func (poolMgr *PoolManagerDaemon) addFeeChunk(chunk *Chunk) {
	log.Printf("adding fee chunk: %v", chunk)
	chunk.status = Reserve
	poolMgr.pool[chunk.addr] = chunk
	poolMgr.retainedAddrs = append(poolMgr.retainedAddrs, chunk.addr)
}

func (poolMgr *PoolManagerDaemon) getReceivableChunks() []btcutil.Address {
	var receivableAddrs []btcutil.Address
	log.Printf("poolmgr: enumerating receivable chunks")
	for addr, chunk := range poolMgr.pool {
		log.Printf("candidate: %v", chunk)
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

func (poolMgr *PoolManagerDaemon) handleBootstrap(bootstrapChunks []*BootstrapFeeChunk) {
	log.Printf("poolmgr bootstrapping chunks")
	for _, bootstrapChunk := range bootstrapChunks {
		chunk, wif, err := bootstrapChunk.normalize()

		log.Printf("importing chunk %v", bootstrapChunk)
		log.Printf("with privkey %v", wif)

		if err != nil {
			log.Printf("error parsing bootstrap chunk: %v", err)
		}

		err = getRpcClient().ImportPrivKey(wif)

		if err != nil {
			log.Printf("error importing privkey: %v", err)
		}

		poolMgr.pool[chunk.addr] = chunk
		poolMgr.retainedAddrs = append(poolMgr.retainedAddrs, chunk.addr)
	}
	log.Printf("retainedAddrs is now: %v", poolMgr.retainedAddrs)
	log.Printf("with length %v", len(poolMgr.retainedAddrs))
}

func (poolMgr *PoolManagerDaemon) prune(height int) {
	var expiredAddrs []string
	for addr, chunk := range poolMgr.pool {
		if chunk.status == Receivable && isExpired(chunk, height) {
			log.Printf("found expired chunk: %s", addr)
			expiredAddrs = append(expiredAddrs, addr)
		}
	}
	for _, addr := range expiredAddrs {
		delete(poolMgr.pool, addr)
	}
}

func (poolMgr *PoolManagerDaemon) handleNew(chunkMsg *ChunkMessage) {
	addr := chunkMsg.MixAddr
	chunk := &Chunk{
		status:  Receivable,
		message: chunkMsg,
		txInfo:  nil,
		addr:    addr,
	}
	poolMgr.pool[addr] = chunk
}

func (poolMgr *PoolManagerDaemon) popRandomMixingChunk() *Chunk {
	log.Printf("generating random index in [0, %d)", len(poolMgr.mixingAddrs))
	randIndex := randInt(len(poolMgr.mixingAddrs))
	randAddr := poolMgr.mixingAddrs[randIndex]
	log.Printf("picked random address %s at index %d", randAddr, randIndex)

	chunk := poolMgr.pool[randAddr]
	log.Printf("popping rand chunk: %v", chunk)
	delete(poolMgr.pool, randAddr)

	// remove from mixingAddrs
	numMixingAddrs := len(poolMgr.mixingAddrs)
	poolMgr.mixingAddrs[randIndex] = poolMgr.mixingAddrs[numMixingAddrs-1]
	poolMgr.mixingAddrs = poolMgr.mixingAddrs[:numMixingAddrs-1]
	return chunk
}

func (poolMgr *PoolManagerDaemon) handleReceived(addr string, txInfo *TxInfo, blockHash *btcwire.ShaHash) {
	// change chunk to mixing, add txoutinfo, add chunk to mixingaddrs,
	// mix
	log.Printf("received chunk at txinfo: %v", txInfo)
	chunk, ok := poolMgr.pool[addr]
	if !ok {
		log.Printf("pool doesn't contain this address: %v", addr)
		return
	}

	chunk.txInfo = txInfo
	log.Printf("assigned txinfo")

	if isFee(chunk.message.Nonce, blockHash, chunk.message.Fee) {
		chunk.status = Reserve
		poolMgr.retainedAddrs = append(poolMgr.retainedAddrs, addr)
		return
	}

	chunk.status = Mixing
	poolMgr.mixingAddrs = append(poolMgr.mixingAddrs, addr)
	log.Printf("added address %s to mixing pool", addr)
	randDelay := generateDelay(chunk.message.ReturnBy)
	log.Printf("generated delay: %v blocks", randDelay)
	outAddr := chunk.message.OutAddr

	go mix(randDelay, outAddr)
}

func (poolMgr *PoolManagerDaemon) popRandomReserveChunk() (*Chunk, error) {
	log.Printf("generating random index in [0, %d)", len(poolMgr.retainedAddrs))
	if len(poolMgr.retainedAddrs) == 0 {
		return nil, errors.New("pool has no retained chunks")
	}
	randIndex := randInt(len(poolMgr.retainedAddrs))
	randAddr := poolMgr.retainedAddrs[randIndex]
	log.Printf("picked random address %s at index %d", randAddr, randIndex)

	chunk := poolMgr.pool[randAddr]
	log.Printf("popping rand chunk: %v", chunk)
	delete(poolMgr.pool, randAddr)
	log.Printf("the chunk is still %v", chunk)

	// remove from retainedAddrs
	numRetainedAddrs := len(poolMgr.retainedAddrs)
	poolMgr.retainedAddrs[randIndex] = poolMgr.retainedAddrs[numRetainedAddrs-1]
	poolMgr.retainedAddrs = poolMgr.retainedAddrs[:numRetainedAddrs-1]
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

func isExpired(chunk *Chunk, height int) bool {
	//currHeight, _ := getBlockchainHeight()
	isPastExpiry := chunk.message.SendBy <= height

	return isPastExpiry
}
