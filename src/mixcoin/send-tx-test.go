package mixcoin

import (
	"btcjson"
	"btcutil"
	"log"
)

func SendChunkTester() {
	chunktest3()
}

func chunktest3() {
	log.Printf("testing chunk sending")

	change := "mzXbdk7wecJRaxQbHcefLSv6GjAmMSUQ8o"
	changeAmount := int64(15000000)

	changeAddr, err := decodeAddress(change)
	if err != nil {
		log.Printf("error decoding address: %v", err)
	}

	dest := "muVYg4SC4fBr5xHR9K2Ym7nb2WcMLSWXPG"
	destAmount := int64(1200000)

	destAddr, err := decodeAddress(dest)
	if err != nil {
		log.Printf("error decoding address: %v", err)
	}

	txInId := "36c1ebe170709f931860e8605a03b74f67814b6d001ada2b3832132556ece5ab"
	txInput := btcjson.TransactionInput{txInId, 0}

	inputs := []btcjson.TransactionInput{txInput}

	amounts := map[btcutil.Address]btcutil.Amount{
		changeAddr: btcutil.Amount(changeAmount),
		destAddr:   btcutil.Amount(destAmount),
	}

	msgTx, err := rpcClient.CreateRawTransaction(inputs, amounts)
	if err != nil {
		log.Printf("error creating tx: %v", err)
	}
	log.Printf("created tx: %v", msgTx)

	signedTx, signed, err := rpcClient.SignRawTransaction(msgTx)
	log.Printf("signed: %v", signed)
	if err != nil {
		log.Printf("error signing tx: %v", err)
		return
	}
	log.Printf("signed tx: %v", signedTx)

	txHash, err := rpcClient.SendRawTransaction(signedTx, true)
	if err != nil {
		log.Printf("error sending tx: %v", err)
		return
	}
	log.Printf("sent tx with tx hash: %v", txHash)
}
