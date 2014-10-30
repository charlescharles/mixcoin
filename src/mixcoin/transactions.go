package mixcoin

import (
	"btcscript"
	"btcwire"
	"log"
)

type TxInfo struct {
	//txOuts         []*btcwire.OutPoint
	receivedAmount int64
	txOut          *btcwire.OutPoint
}

func sendChunk(chunk *Chunk, dest string) error {
	log.Printf("sending a chunk to %s", dest)
	tx := btcwire.NewMsgTx()

	/**
	for _, prevOut := range chunk.txInfo.txOuts {
		tx.AddTxIn(btcwire.NewTxIn(prevOut, make([]byte, 10)))
	}
	*/

	tx.AddTxIn(&btcwire.NewTxIn(chunk.txInfo.txOut), make([]byte, 10))

	destAddr, err := decodeAddress(dest)
	if err != nil {
		log.Panicln("error decoding address: %v", err)
	}
	pkScript, err := btcscript.PayToAddrScript(destAddr)
	if err != nil {
		log.Panicf("error creating pkscript: %v", err)
		return err
	}

	txOut := btcwire.NewTxOut(chunk.txInfo.receivedAmount, pkScript)
	tx.AddTxOut(txOut)

	tx, signed, err := rpcClient.SignRawTransaction(tx)
	if !signed {
		log.Printf("couldn't sign input transaction!")
	}
	if err != nil {
		log.Panicf("error signing input transaction: %v", err)
	}

	// allow high fees?
	txHash, err := rpcClient.SendRawTransaction(tx, true)

	if err != nil {
		log.Panicf("error sending transaction: %v", err)
		return err
	}

	log.Printf("sent transaction with hash: %v", txHash)
	return nil
}
