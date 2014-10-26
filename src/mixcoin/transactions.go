package mixcoin

import (
	"btcscript"
	"btcwire"
	"log"
)

type TxInfo struct {
	txOuts         []*btcwire.OutPoint
	receivedAmount float64
}

func sendChunk(chunk *Chunk, dest string) error {
	builder := btcscript.NewScriptBuilder()
	tx := btcwire.NewMsgTx()

	for _, prevOut := range chunk.txInfo.txOuts {
		tx.AddTxIn(btcwire.NewTxIn(prevOut, make([]byte, 10)))
	}

	destAddr, err := decodeAddress(dest)
	if err != nil {
		log.Panicln("error decoding address")
	}
	pkScript, err := btcscript.PayToAddrScript(destAddr)
	if err != nil {
		log.Panicf("error creating pkscript: ", err)
		return err
	}

	txOut := btcwire.NewTxOut(chunk.txInfo.receivedAmount, pkScript)
	tx.AddTxOut(txOut)

	tx, signed, err := rpcClient.SignRawTransaction(tx)
	if !signed || err != nil {
		log.Panicf("error signing input transactions: ", err)
		return err
	}

	// allow high fees?
	txHash, err := rpcClient.SendRawTransaction(tx, true)

	if err != nil {
		log.Panicf("error sending transaction: ", err)
		return err
	}
	return nil
}
