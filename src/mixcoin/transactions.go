package mixcoin

import (
	"github.com/conformal/btcjson"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"log"
)

type TxInfo struct {
	receivedAmount int64
	txOut          *btcwire.OutPoint
}

func sendChunkWithFee(inputChunk *Chunk, dest string) error {
	log.Printf("sending the following chunk to %s:", dest)
	log.Printf("%v", inputChunk)

	cfg := GetConfig()

	feeChunk := GetPool().GetRandomChunk(Reserve)
	feeChunkAmt := feeChunk.txInfo.receivedAmount
	feeTxOut, err := feeChunk.GetAsTxInput()

	if err != nil {
		log.Printf("error getting txinput: %v", err)
	}

	inputChunkAmt := inputChunk.txInfo.receivedAmount
	inputTxOut, err := inputChunk.GetAsTxInput()

	if err != nil {
		log.Printf("error getting txinput: %v", err)
	}

	destAmt := cfg.ChunkSize
	destAddr, err := decodeAddress(dest)
	if err != nil {
		log.Printf("error decoding address: %v", err)
	}

	changeAddr, err := decodeAddress(feeChunk.addr)
	if err != nil {
		log.Printf("error decoding address: %v", err)
	}
	changeAmt := feeChunkAmt + inputChunkAmt - destAmt - cfg.TxFee

	inputs := []btcjson.TransactionInput{*feeTxOut, *inputTxOut}

	outAmounts := map[btcutil.Address]btcutil.Amount{
		destAddr:   btcutil.Amount(destAmt),
		changeAddr: btcutil.Amount(changeAmt),
	}

	msgTx, err := getRpcClient().CreateRawTransaction(inputs, outAmounts)
	if err != nil {
		log.Printf("error creating tx: %v", err)
	}
	log.Printf("created tx: %v", msgTx)

	signedTx, signed, err := getRpcClient().SignRawTransaction(msgTx)
	log.Printf("signed: %v", signed)
	if err != nil {
		log.Printf("error signing tx: %v", err)
		return err
	}
	log.Printf("signed tx: %v", signedTx)

	txHash, err := getRpcClient().SendRawTransaction(signedTx, true)
	if err != nil {
		log.Printf("error sending tx: %v", err)
		return err
	}
	log.Printf("sent tx with tx hash: %v", txHash)

	feeChunk.txInfo.receivedAmount -= cfg.TxFee
	if feeChunk.txInfo <= 0 {
		log.Printf("used up fee chunk")
	} else {
		log.Printf("adding fee chunk back to pool: %v", feeChunk)
		addFeeChunkToPool(feeChunk)
	}

	return nil
}
