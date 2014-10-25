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
}

func handleChunkRequest(chunk *Chunk) (*Chunk, error) {
	log.Println("handling chunk request")

	addr, err := getNewAddress()
	if err != nil {
		log.Panicf("Unable to create new address: " + err)
	}

	encodedAddr := (*addr).EncodeAddress()

	chunk.EscrowAddr = encodedAddr

	err = signChunk(chunk)
	if err != nil {
		log.Panicf("Couldn't sign chunk: " + err)
	}

	registerNewChunk(encodedAddr, chunk)
	registerAddress(encodedAddr)

	return chunk, nil
}

func getNewAddress() (*btcutil.Address, error) {
	addr, err := rpcClient.GetNewAddress()
	if err != nil {
		rpcClient.CreateEncryptedWallet("Thereis1")
	}
	addr, err = rpcClient.GetNewAddress()
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

func onNewBlock(blockHash *btcwire.ShaHash, height int32) {
	cfg := GetConfig()

	minConf := cfg.MinConfirmations
	stdChunkSize := cfg.ChunkSize

	blockVerbose, err := rpcClient.GetBlockVerbose(blockHash, true)

	// for x, rawTx := range blockVerbose.RawTx {
	// 	for voutIndex, vout := range rawTx.Vout {
	// 		if vout.Value >= stdChunkSize && len()
	// 	}
	// }

	for encodedAddr, chunk := range poolManager.Receivable.lookup {
		addr := decodeAddress(encodedAddr)
		amount, err := rpcClient.GetReceivedByAddressMinConf(addr, minConf)
		if err != nil {
			return err
		}
		account, err := rpcClient.GetAccount(addr)
		if err != nil {
			return err
		}
		// TODO check that the time is before receivedBy
		if amount >= stdChunkSize {
			// check random beacon to see if we should retain as fee
			// move chunk from receivable into pool
		}
	}
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
