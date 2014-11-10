package mixcoin

import (
	"log"

	"github.com/conformal/btcjson"
	"github.com/conformal/btcutil"
)

func send(dest string) error {
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

	feeTx := &btcjson.TransactionInput{feeUtxo.TxId, uint32(feeUtxo.Index)}
	inputTx := &btcjson.TransactionInput{inputUtxo.TxId, uint32(inputUtxo.Index)}

	changeAddr, err := decodeAddress(feeUtxo.Addr)
	if err != nil {
		log.Printf("error decoding change Address: %v", err)
	}
	destAddr, err := decodeAddress(dest)
	if err != nil {
		log.Printf("error decoding change Address: %v", err)
	}

	feeAmt := btcutil.Amount(cfg.TxFee)
	destAmt := btcutil.Amount(cfg.ChunkSize)
	changeAmt := feeUtxo.Amount + inputUtxo.Amount - destAmt - feeAmt

	inputs := []btcjson.TransactionInput{*feeTx, *inputTx}

	amounts := map[btcutil.Address]btcutil.Amount{
		destAddr:   destAmt,
		changeAddr: changeAmt,
	}

	log.Printf("creating raw tx with inputs: %+v\namounts:%+v", inputs, amounts)
	tx, err := rpc.CreateRawTransaction(inputs, amounts)
	if err != nil {
		log.Printf("error creating tx: %v", err)
	}

	log.Printf("created tx: %+v", tx)

	signed, ok, err := rpc.SignRawTransaction(tx)
	log.Printf("signed: %v", ok)
	if err != nil {
		log.Printf("error signing tx: %v", err)
	}
	log.Printf("signed tx: %v", signed)

	txHash, err := rpc.SendRawTransaction(signed, true)
	if err != nil {
		log.Printf("error sending tx: %+v", err)
	}
	log.Printf("sent tx with tx hash: %v", txHash)

	feeUtxo.Amount -= feeAmt
	feeUtxo.TxId = txHash.String()
	if feeUtxo.Amount <= feeAmt {
		log.Printf("used up fee chunk")
	} else {
		log.Printf("adding fee chunk back to pool")
		pool.Put(Reserve, feeUtxo)
	}
	return nil
}
