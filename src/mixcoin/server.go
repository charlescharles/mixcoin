package mixcoin

import (
	"btcjson"
	"btcutil"
	"btcwire"
	"errors"
	"log"
)

const (
	MAX_CONF = 9999
)

func StartMixcoinServer() {
	StartRpcClient()
	StartPoolManager()
}

func handleChunkRequest(chunkMsg *ChunkMessage) error {
	log.Println("handling chunk request")

	err := validateChunkMsg(chunkMsg)
	if err != nil {
		log.Printf("Invalid chunk request")
		return err
	}

	addr, err := getNewAddress()
	if err != nil {
		log.Panicf("Unable to create new address")
		return err
	}

	encodedAddr := (*addr).EncodeAddress()

	chunkMsg.MixAddr = encodedAddr

	err = signChunkMessage(chunkMsg)
	if err != nil {
		log.Panicf("Couldn't sign chunk")
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
	return nil
}

func registerNewChunk(encodedAddr string, chunkMsg *ChunkMessage) {
	newChunkC <- &NewChunk{encodedAddr, chunkMsg}
}

func onNewBlock(blockHash *btcwire.ShaHash, height int32) {
	cfg := GetConfig()
	minConf := cfg.MinConfirmations
	stdChunkSize := cfg.ChunkSize

	// TODO don't access the pool
	var receivableAddrs []btcutil.Address
	for addr, chunk := range pool {
		if chunk.status == Receivable {
			decoded, err := decodeAddress(addr)
			if err != nil {
				log.Panicln("unable to decode address")
			}
			receivableAddrs = append(receivableAddrs, decoded)
		}
	}

	receivedByAddress, err := rpcClient.ListUnspentMinMaxAddresses(minConf, MAX_CONF, receivableAddrs)
	if err != nil {
		log.Panicf("error listing unspent by address: ", err)
	}
	received := make(map[string]*TxInfo)
	for _, result := range receivedByAddress {
		addr := result.Address

		txInfo, exists := received[addr]
		if !exists {
			received[addr] = &TxInfo{}
			txInfo = received[addr]
		}

		txHash, err := btcwire.NewShaHashFromStr(result.TxId)
		outpoint := &btcwire.OutPoint{
			*txHash,
			result.Vout,
		}

		txInfo.receivedAmount += result.Amount
		txInfo.txOuts = append(txInfo.txOuts, outpoint)

		receivedChunkC <- &ReceivedChunk{addr, txInfo}
	}
}

func isValidReceivedResult(result *btcjson.ListUnspentResult) bool {
	cfg := GetConfig()

	hasConfirmations := result.Confirmations >= cfg.MinConfirmations
	hasAmount := result.Amount >= cfg.ChunkSize

	return hasConfirmations && hasAmount
}
