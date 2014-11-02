package mixcoin

import (
	"btcjson"
	"btcutil"
	"btcwire"
	"log"
)

type TxInfo struct {
	receivedAmount int64
	txOut          *btcwire.OutPoint
}

func sendChunk(chunk *Chunk, dest string) error {
	log.Printf("sending the following chunk to %s:", dest)
	log.Printf("%v", chunk)

	txInfo := chunk.txInfo
	destAmount := txInfo.receivedAmount - 200000
	destAddr, err := decodeAddress(dest)
	if err != nil {
		log.Printf("error decoding address: %v", err)
	}

	txInput := btcjson.TransactionInput{txInfo.txOut.Hash.String(), txInfo.txOut.Index}

	inputs := []btcjson.TransactionInput{txInput}

	outAmounts := map[btcutil.Address]btcutil.Amount{
		destAddr: btcutil.Amount(destAmount),
	}

	msgTx, err := rpcClient.CreateRawTransaction(inputs, outAmounts)
	if err != nil {
		log.Printf("error creating tx: %v", err)
	}
	log.Printf("created tx: %v", msgTx)

	signedTx, signed, err := rpcClient.SignRawTransaction(msgTx)
	log.Printf("signed: %v", signed)
	if err != nil {
		log.Printf("error signing tx: %v", err)
		return err
	}
	log.Printf("signed tx: %v", signedTx)

	txHash, err := rpcClient.SendRawTransaction(signedTx, true)
	if err != nil {
		log.Printf("error sending tx: %v", err)
		return err
	}
	log.Printf("sent tx with tx hash: %v", txHash)

	return nil
}
