package mixcoin

import (
	"btcutil"
	"btcwire"
	"log"
)

type BootstrapMixChunk struct {
	addr   string
	amount float64
	txId   string
	index  int
}

func (bootstrapMixChunk *BootstrapMixChunk) toReceivable() (*ReceivedChunk, error) {
	txHash, err := btcwire.NewShaHashFromStr(bootstrapMixChunk.txId)
	if err != nil {
		log.Printf("error creating sha hash from bootstrap txid: %v", err)
	}
	outpoint := btcwire.NewOutPoint(txHash, uint32(bootstrapMixChunk.index))
	amountSatoshi := int64(bootstrapMixChunk.amount * btcutil.SatoshiPerBitcoin)

	txInfo := &TxInfo{
		receivedAmount: amountSatoshi,
		txOut:          outpoint,
	}

	recvd := &ReceivedChunk{
		addr:   bootstrapMixChunk.addr,
		txInfo: txInfo,
	}
	return recvd, nil
}

var (
	mixPool = []*BootstrapMixChunk{
		&BootstrapMixChunk{
			addr:   "n4hESCEbYLgiZURGYXngzMhHcdSyWbNqTj",
			amount: 8.2,
			txId:   "a43ccc3b34ded6f18ccb5c066e421148a67334131f58067bbfc74e66a20d98b3",
			index:  0,
		},
	}
)

func BootstrapMixingPool() {
	log.Printf("bootstrapping mix pool")
	//bootstrapMixC <- mixPool
}
