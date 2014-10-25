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

type Pool struct {
	mutex  sync.RWMutex
	lookup map[string]*Chunk
}

func (pool *Pool) Add(addr string, chunk *Chunk) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.lookup[addr] = chunk
	return nil
}

func (pool *Pool) Remove(addr string) (*Chunk, error) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	chunk, exists := pool.lookup[addr]
	if !exists {
		return nil, Error("chunk at address ", addr, " doesn't exist")
	}

	delete(pool, addr)

	return chunk, nil
}

func (pool *Pool) Get(addr string) (*Chunk, error) {
	pool.mutex.RLock()
	defer pool.mutex.RUnlock()

	chunk, exists := pool.lookup[addr]
	if !exists {
		return nil, Error("chunk at address ", addr, " doesn't exist")
	}
	return chunk, nil
}

type PoolManager struct {
	Receivable *Pool
	Mixing     *Pool
	Retained   *Pool

	AddrToType           map[string]PoolType
	MixingChunkAddresses []string
}

var (
	poolManager PoolManager
	poolForType map[PoolType]*Pool
)

func init() {
	poolManager = PoolManager{}

	poolForType = map[PoolType]*Pool{
		Receivable: poolManager.Receivable,
		Mixing:     poolManager.Mixing,
		Retained:   poolManager.Retained,
	}
}

/**
* Atomically add a chunk to this pool
 */
func AddChunk(addr string, chunk *Chunk, poolType PoolType) error {
	pool := poolForType[poolType]
	return pool.Add(addr, chunk)
}

/**
* Atomically move chunk
 */
func MoveChunk(addr string, source, dest PoolType) error {
	sourcePool, err := poolForType[source]
	if err != nil {
		return err
	}

	destPool, err := poolForType[dest]
	if err != nil {
		return err
	}

	sourcePool.mutex.Lock()
	destPool.mutex.Lock()
	defer sourcePool.mutex.Unlock()
	defer destPool.mutex.Unlock()

	chunk, exists = sourcePool.lookup[addr]
	if !exists {
		return Error("chunk at address ", addr, " doesn't exist")
	}
	destPool.lookup[addr] = chunk
	delete(sourcePool.lookup, addr)
}

/**
* Atomically remove and return chunk with given address and pool
 */
func PopChunk(addr string, poolType PoolType) (*Chunk, error) {
	pool := poolForType[poolType]
	chunk, err := pool.Remove(addr)
	if err != nil {
		return nil, err
	}

	return chunk, nil
}

/**
* Atomically remove and return a random chunk in mixing pool
 */
func PopRandomMixingChunk() (*Chunk, error) {
	mixingAddressCount := len(poolManager.mixingChunkAddresses)
	randAddress := poolManager.mixingChunkAddresses[rand.Intn(mixingAddressCount)]
	chunk, err := PopChunk(randAddress, Mixing)
	if err != nil {
		panic(err)
	}

	return chunk, nil
}
