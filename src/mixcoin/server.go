package mixcoin

import (
	"btcnet"
	"btcrpcclient"
	"btcutil"
	"btcwire"
	"btcws"

	"btcjson"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const (
	MAX_CONF = 9999
)

var (
	rpcClient *btcrpcclient.Client
)

func StartMixcoinServer() {
	cfg := GetConfig()

	log.Printf("Reading btcd cert file %s", cfg.CertFile)
	certs, err := ioutil.ReadFile(cfg.CertFile)
	if err != nil {
		log.Panicf("couldn't read btcd certs")
	}

	connCfg := &btcrpcclient.ConnConfig{
		Host:         cfg.RpcAddress,
		Endpoint:     "ws",
		User:         cfg.RpcUser,
		Pass:         cfg.RpcPass,
		Certificates: certs,
		DisableTLS:   false,
	}

	ntfnHandlers := btcrpcclient.NotificationHandlers{
		OnBlockConnected: OnNewBlock,
		OnRecvTx:         OnReceivedTx,
	}

	client, err := btcrpcclient.New(connCfg, &ntfnHandlers)
	if err != nil {
		log.Panicf(err)
	}

	// Register for block connect and disconnect notifications.
	if err = client.NotifyBlocks(); err != nil {
		log.Panicf(err)
	}

	StartPoolManager()
}

func handleChunkRequest(chunk *Chunk) (*Chunk, error) {
	log.Println("handling chunk request")

	addr, err := getNewAddress()
	if err != nil {
		log.Panicf("Unable to create new address: " + err)
	}

	encodedAddr := (*addr).EncodeAddress()

	chunk.MixAddr = encodedAddr

	err = signChunk(chunk)
	if err != nil {
		log.Panicf("Couldn't sign chunk: " + err)
	}

	registerNewChunk(encodedAddr, chunk)
	registerAddress(encodedAddr)

	return chunk, nil
}

func getNewAddress() (*btcutil.Address, error) {
	cfg := GetConfig()

	addr, err := rpcClient.GetNewAddress()
	if err != nil {
		rpcClient.CreateEncryptedWallet(cfg.WalletPass)
		addr, err = rpcClient.GetNewAddress()
	}
	if err != nil {
		log.Panicf(err)
		return nil, err
	}
	err = rpcClient.SetAccount(addr, cfg.MixAccount)
	if err != nil {
		log.Panicf(err)
		return nil, err
	}
	return &addr, nil
}

func registerNewChunk(encodedAddr string, chunk *Chunk) error {
	return nil
}

func registerAddress(encoded string) error {
	addr, err := decodeAddress(encoded)

	rpcClient.NotifyReceived([]*btcutil.Address{&addr})
	return nil
}

func onReceivedTx(tx *btcutil.Tx, details *btcws.BlockDetails) {
	blockHash := details.Hash
	return
}

func extractReceivedVouts(tx *btcjson.TxRawResult) (map[string]*btcjson.Vout, error) {
	ret = make(map[string]*btcjson.Vout)

	cfg := GetConfig()

	for voutIndex, vout := range tx.Vout {
		lockScript := vout.ScriptPubKey
		lock
		if vout.Value >= cfg.ChunkSize &&
			lockScript.Type == "scriptpubkey" &&
			len(lockScript.addresses == 1) && lockScript.addresses[0] == "one of the receivable addresses" {
			ret[lockScript.addresses[0].EncodeAddress()] = vout
		}
	}

	return ret, nil
}

func pruneExpiredChunks() ([]string, error) {
	poolManager.Receivable.mutex.Lock()
	defer poolManager.Receivable.mutex.Unlock()

	addrs := make([]string)

	for addr, chunk := range poolManager.Receivable.chunkMap {
		if isChunkExpired(chunk) {
			delete(poolManager.Receivable.chunkMap, addr)
		} else {
			addrs = append(addrs, addr)
		}
	}

	return addrs, nil
}

func isChunkExpired(chunk *Chunk) bool {
	return false
}

func onNewBlock(blockHash *btcwire.ShaHash, height int32) {
	receivableAddrs, err := pruneExpiredChunks()
	if err != nil {
		log.Panicln("error pruning expired chunks: ", err)
	}

	cfg := GetConfig()
	minConf := cfg.MinConfirmations
	stdChunkSize := cfg.ChunkSize

	receivedByAddress, err := rpcClient.ListUnspentMinMaxAddresses(minConf, MAX_CONF, receivableAddrs)
	if err != nil {
		log.Panicf("error listing unspent by address: ", err)
	}

	receivePool := poolManager.Receivable
	receivedChunkAddrs := make([]string)

	for _, result := range receivedByAddress {
		addr := result.Address

		receivePool.mutex.Lock()
		outpoints := receivePool.chunkMap[addr].TxInfo.txOuts

		receivePool.chunkMap[addr].TxInfo.receivedAmount += result.Amount

		txHash := btcwire.NewShaHashFromStr(result.TxId)

		outpoints = append(outpoints, &btcwire.Outpoint{
			txHash,
			result.Vout,
		})

		receivedChunkAddrs = append(receivedChunkAddrs, addr)
		receivePool.mutex.Unlock()
	}
}

func isValidReceivedResult(result *btcjson.ListUnspentResult) bool {
	cfg := GetConfig()

	hasConfirmations := result.Confirmations >= cfg.MinConfirmations
	hasAmount := result.Amount >= cfg.ChunkSize

	return hasConfirmations && hasAmount
}

func handleReceivedChunk(addr string, txOutInfo *TxOutInfo) error {
	chunk, err := poolManager.Receivable.Get(addr)
	if err != nil {
		return err
	}
	chunk.TxOut = txOutInfo
	MoveChunk(addr, Receivable, Mixing)

	go mix(chunk)
}
