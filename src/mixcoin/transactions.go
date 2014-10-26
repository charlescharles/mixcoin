package mixcoin

import (
	"btcscript"
	"btcwire"
)

type TxInfo struct {
	txOuts         []*btcwire.OutPoint
	receivedAmount float64
}

func sendChunk(chunk *Chunk, dest string) error {
	builder := btcscript.NewScriptBuilder()
	tx := btcwire.NewMsgTx()

	for _, prevOut := range chunk.TxInfo.txOuts {
		tx.AddTxIn(btcwire.NewTxIn(prevOut, []byte))
	}

	destAddr := decodeAddress(dest)
	pkScript, err := btcscript.PayToAddrScript(destAddr)
	if err != nil {
		log.Panicf("error creating pkscript: ", err)
		return err
	}

	txOut, err := btcwire.NewTxOut(chunk.TxInfo.receivedAmount, pkScript)
	tx.AddTxOut(txOut)

	tx, signed, err := rpcClient.SignRawTransaction(tx)
	if !signed || err != nil {
		log.Panicf("error signing input transactions: ", err)
		return err
	}

	// allow high fees?
	txHash, err = rpcClient.SendRawTransaction(tx, true)

	if err != nil {
		log.Panicf("error sending transaction: ", err)
		return err
	}
}
