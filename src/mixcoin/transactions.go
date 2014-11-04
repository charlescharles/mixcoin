package mixcoin

import (
	"github.com/conformal/btcjson"
	"github.com/conformal/btcutil"
	"log"
)

func send(dest string) error {
	cfg := GetConfig()

	feeItem, err := pool.Get(Reserve)
	feeUtxo := feeItem.(*Utxo)
	if err != nil {
		log.Printf("error getting input utxo: %v", err)
	}

	inputItem, err := pool.Get(Mixing)
	inputUtxo := inputItem.(*Utxo)
	if err != nil {
		log.Printf("error getting input utxo: %v", err)
	}

	feeTx := &btcjson.TransactionInput{feeUtxo.txId, uint32(feeUtxo.index)}
	inputTx := &btcjson.TransactionInput{inputUtxo.txId, uint32(inputUtxo.index)}

	changeAddr, err := decodeAddress(feeUtxo.addr)
	if err != nil {
		log.Printf("error decoding change address: %v", err)
	}
	destAddr, err := decodeAddress(dest)
	if err != nil {
		log.Printf("error decoding change address: %v", err)
	}

	feeAmt := btcutil.Amount(cfg.TxFee)
	destAmt := btcutil.Amount(cfg.ChunkSize)
	changeAmt := feeUtxo.amount + inputUtxo.amount - destAmt - feeAmt

	inputs := []btcjson.TransactionInput{*feeTx, *inputTx}

	amounts := map[btcutil.Address]btcutil.Amount{
		destAddr:   destAmt,
		changeAddr: changeAmt,
	}

	tx, err := rpc.CreateRawTransaction(inputs, amounts)
	if err != nil {
		log.Printf("error creating tx: %v", err)
	}
	log.Printf("created tx: %v", tx)

	signed, ok, err := rpc.SignRawTransaction(tx)
	log.Printf("signed: %v", ok)
	if err != nil {
		log.Printf("error signing tx: %v", err)
	}
	log.Printf("signed tx: %v", signed)

	txHash, err := rpc.SendRawTransaction(signed, true)
	if err != nil {
		log.Printf("error sending tx: %v", err)
	}
	log.Printf("sent tx with tx hash: %v", txHash)

	feeUtxo.amount -= feeAmt
	if feeUtxo.amount <= feeAmt {
		log.Printf("used up fee chunk")
	} else {
		log.Printf("adding fee chunk back to pool")
		pool.Put(Reserve, feeUtxo)
	}
	return nil
}
