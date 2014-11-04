package mixcoin

import (
	"errors"
	"github.com/conformal/btcjson"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"log"
)

const (
	MAX_CONF = 9999
)

var (
	blockchainHeight int
	pool             *PoolManager
	rpc              RpcClient
	mix              *Mix
)

func StartMixcoinServer() {
	log.Println("starting mixcoin server")

	pool = NewPoolManager()
	rpc = NewRpcClient()
	mix = NewMix()

	BootstrapPool()
}

func handleChunkRequest(chunkMsg *ChunkMessage) error {
	log.Printf("handling chunk request: %s", chunkMsg)

	err := validateChunkMsg(chunkMsg)
	if err != nil {
		log.Printf("Invalid chunk request: %v", err)
		return err
	}

	log.Printf("generating new address")
	addr, err := getNewAddress()
	if err != nil {
		log.Panicf("Unable to create new address: %v", err)
		return err
	}

	encodedAddr := (*addr).EncodeAddress()
	log.Printf("generated address: %s", encodedAddr)

	chunkMsg.MixAddr = encodedAddr

	err = signChunkMessage(chunkMsg)
	if err != nil {
		log.Panicf("Couldn't sign chunk: %v", err)
		return err
	}

	registerNewChunk(encodedAddr, chunkMsg)
	return nil
}

func registerNewChunk(encodedAddr string, chunkMsg *ChunkMessage) {
	log.Printf("registering new chunk at address %s", encodedAddr)
	pool.Put(Receivable, chunkMsg)
	log.Printf("added chunk to pool")
	decoded, _ := decodeAddress(encodedAddr)
	log.Printf("set notification for address %s", decoded)
}

func onBlockConnected(blockHash *btcwire.ShaHash, height int32) {
	log.Printf("new block connected with hash %v, height %d", blockHash, height)

	blockchainHeight = int(height)
	go findTransactions(blockHash, int(height))
}

func prune() {
	pool.Filter(func(msg *ChunkMessage) bool {
		return msg.SendBy <= blockchainHe
	})
}

func findTransactions(blockHash *btcwire.ShaHash, height int) {
	prune()

	cfg := GetConfig()
	minConf := cfg.MinConfirmations

	log.Printf("getting receivable chunks")
	receivableAddrs := pool.ReceivingKeys()
	log.Printf("current receivable addresses: %v", receivableAddrs)

	receivedByAddress, err := rpc.ListUnspentMinMaxAddresses(minConf, MAX_CONF, receivableAddrs)
	if err != nil {
		log.Panicf("error listing unspent by address: %v", err)
	}
	log.Printf("received transactions: %v", receivedByAddress)

	// make addr -> utxo map of received txs
	received := make(map[string]*TxInfo)
	for _, result := range receivedByAddress {
		if !isValidReceivedResult(result) {
			continue
		}

		amount, err := btcutil.NewAmount(result.Amount)
		if err != nil {
			log.Panicf("invalid tx amount: %v", err)
		}

		received[addr] = &Utxo{
			addr:   result.Address,
			amount: amount,
			txId:   result.TxId,
			index:  int(result.Vout),
		}
	}

	var receivedAddrs []string
	for addr, _ := range received {
		receivedAddrs = append(receivedAddrs, addr)
	}

	// get the chunk messages, move to pool
	chunkMsgs := pool.Scan(receivedAddrs)
	for _, msg := range chunkMsgs {
		utxo := received[msg.MixAddr]
		if isFee(msg.Nonce, blockHash, msg.Fee) {
			pool.Put(Reserve, utxo)
		} else {
			pool.Put(Mixing, utxo)
			mix.Put(msg)
		}
	}
	log.Printf("done handling block")
}

func isFee(nonce int64, hash *btcwire.ShaHash, feeBips int) bool {
	bigIntHash := big.NewInt(0)
	bigIntHash.SetBytes(hash.Bytes())
	hashInt := bigIntHash.Int64()

	gen := nonce | hashInt
	fee := float64(feeBips) * 1.0e4

	source := rand.NewSource(gen)
	rng := rand.New(source)
	return rng.Float64() <= fee
}

func isValidReceivedResult(result *btcjson.ListUnspentResult) bool {
	cfg := GetConfig()

	// ListUnspentResult.Amount is a float64 in BTC
	// btcutil.Amount is an int64
	amountReceived, err := btcutil.NewAmount(result.Amount)
	if err != nil {
		log.Printf("error parsing amount received: %v", err)
	}
	amountReceivedInt := int64(amountReceived)

	hasConfirmations := result.Confirmations >= int64(cfg.MinConfirmations)
	hasAmount := amountReceivedInt >= cfg.ChunkSize

	return hasConfirmations && hasAmount
}
