package mixcoin

import (
	"btcjson"
	"btcutil"
	"btcwire"
	"btcws"
	"errors"
	"log"
)

const (
	MAX_CONF = 9999
)

func StartMixcoinServer() {
	log.Println("starting mixcoin server")

	StartRpcClient()
	StartPoolManager()
}

func handleChunkRequest(chunkMsg *ChunkMessage) error {
	log.Printf("handling chunk request: %s", chunkMsg)

	err := validateChunkMsg(chunkMsg)
	if err != nil {
		log.Printf("Invalid chunk request: %v", err)
		return err
	}

	log.Printf("validated chunk request")

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

	currHeight, err := getBlockchainHeight()
	if err != nil {
		return err
	}
	if chunkMsg.SendBy-currHeight > cfg.MaxFutureChunkTime {
		return errors.New("sendby time too far in the future")
	}
	if chunkMsg.SendBy <= currHeight {
		return errors.New("sendby time has already passed")
	}
	if chunkMsg.ReturnBy-chunkMsg.SendBy < 2 {
		return errors.New("not enough time between sendby and returnby")
	}
	return nil
}

func registerNewChunk(encodedAddr string, chunkMsg *ChunkMessage) {
	newChunkC <- &NewChunk{encodedAddr, chunkMsg}
	decoded, _ := decodeAddress(encodedAddr)
	log.Printf("set notification for address %s", decoded)
	rpcClient.NotifyReceived([]btcutil.Address{decoded})
}

func onRecvTx(transaction *btcutil.Tx, details *btcws.BlockDetails) {
	log.Printf("received transaction: %v", transaction)
	log.Printf("block details: %v", details)
}

func onBlockConnected(blockHash *btcwire.ShaHash, height int32) {
	log.Printf("new block connected with hash %v, height %d", blockHash, height)

	go findTransactions()
}

func findTransactions() {
	cfg := GetConfig()
	minConf := cfg.MinConfirmations

	// TODO don't access the pool
	var receivableAddrs []btcutil.Address
	for addr, chunk := range pool {
		if chunk.status == Receivable {
			decoded, err := decodeAddress(addr)
			if err != nil {
				log.Panicf("unable to decode address: %v", err)
			}
			receivableAddrs = append(receivableAddrs, decoded)
		}
	}
	log.Printf("current receivable addresses: %v", receivableAddrs)

	receivedByAddress, err := rpcClient.ListUnspentMinMaxAddresses(minConf, MAX_CONF, receivableAddrs)
	if err != nil {
		log.Panicf("error listing unspent by address: %v", err)
	}
	log.Printf("received transactions: %v", receivedByAddress)
	received := make(map[string]*TxInfo)
	for _, result := range receivedByAddress {
		addr := result.Address

		txInfo, exists := received[addr]
		if !exists {
			received[addr] = &TxInfo{}
			txInfo = received[addr]
		}

		txHash, err := btcwire.NewShaHashFromStr(result.TxId)
		if err != nil {
			log.Printf("error parsing tx sha hash: %v", err)
		}
		outpoint := &btcwire.OutPoint{
			*txHash,
			result.Vout,
		}
		// result.Amount is float64
		received, err := btcutil.NewAmount(result.Amount)
		txInfo.receivedAmount = int64(received)
		txInfo.txOut = outpoint

		receivedChunkC <- &ReceivedChunk{addr, txInfo}
	}
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
