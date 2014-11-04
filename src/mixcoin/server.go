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
)

func StartMixcoinServer() {
	log.Println("starting mixcoin server")

	pool = NewPoolManager()

	StartRpcClient()
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

func validateChunkMsg(chunkMsg *ChunkMessage) error {
	cfg := GetConfig()

	if chunkMsg.Val != cfg.ChunkSize {
		return errors.New("Invalid chunk size")
	}
	if chunkMsg.Confirm < cfg.MinConfirmations {
		return errors.New("Invalid number of confirmations")
	}

	height, err := getBlockchainHeight()
	if err != nil {
		return err
	}
	blockchainHeight = height

	if chunkMsg.SendBy-blockchainHeight > cfg.MaxFutureChunkTime {
		return errors.New("sendby time too far in the future")
	}
	if chunkMsg.SendBy <= blockchainHeight {
		return errors.New("sendby time has already passed")
	}
	if chunkMsg.ReturnBy-chunkMsg.SendBy < 2 {
		return errors.New("not enough time between sendby and returnby")
	}
	log.Printf("validated block")
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

func findTransactions(blockHash *btcwire.ShaHash, height int) {
	GetPool().Prune(height)

	cfg := GetConfig()
	minConf := cfg.MinConfirmations

	log.Printf("getting receivable chunks")
	receivableAddrs := pool.ReceivingKeys()
	log.Printf("current receivable addresses: %v", receivableAddrs)

	receivedByAddress, err := getRpcClient().ListUnspentMinMaxAddresses(minConf, MAX_CONF, receivableAddrs)
	if err != nil {
		log.Panicf("error listing unspent by address: %v", err)
	}
	log.Printf("received transactions: %v", receivedByAddress)

	// make addr -> utxo map of received txs
	received := make(map[string]*TxInfo)
	for _, result := range receivedByAddress {
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

	// get the chunk messages
	chunkMsgs := pool.
		log.Printf("done handling block")
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
