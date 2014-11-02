package mixcoin

import (
	"btcjson"
	"btcutil"
	"log"
)

func SendChunkTestnet() {
	chunktest2()
}

func chunktest2() {
	log.Printf("testing chunk sending")

	origin := "n4hESCEbYLgiZURGYXngzMhHcdSyWbNqTj"
	originAmount := int64(8.2 * btcutil.SatoshiPerBitcoin)

	log.Printf("originAmount: %v", originAmount)

	change := origin
	changeAmount := int64(8.07 * btcutil.SatoshiPerBitcoin)

	changeAddr, _ := decodeAddress(change)

	receivedByAddress, err := rpcClient.ListUnspentMinMaxAddresses(0, 9999, []btcutil.Address{changeAddr})
	log.Printf("received %v", receivedByAddress)

	dest := "muVYg4SC4fBr5xHR9K2Ym7nb2WcMLSWXPG"
	destAmount := int64(0.02 * btcutil.SatoshiPerBitcoin)

	destAddr, _ := decodeAddress(dest)

	txInId := "a43ccc3b34ded6f18ccb5c066e421148a67334131f58067bbfc74e66a20d98b3"
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
