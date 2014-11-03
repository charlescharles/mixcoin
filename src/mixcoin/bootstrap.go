package mixcoin

import (
	"btcutil"
	"btcwire"
	"log"
)

type BootstrapFeeChunk struct {
	addr       string
	amount     float64
	txId       string
	index      int
	privKeyWif string
}

func (bootstrapChunk *BootstrapFeeChunk) normalize() (*Chunk, *btcutil.WIF, error) {
	txHash, err := btcwire.NewShaHashFromStr(bootstrapChunk.txId)
	if err != nil {
		log.Printf("error creating sha hash from bootstrap txid: %v", err)
		return nil, nil, err
	}
	outpoint := btcwire.NewOutPoint(txHash, uint32(bootstrapChunk.index))
	amountSatoshi := int64(bootstrapChunk.amount * btcutil.SatoshiPerBitcoin)

	txInfo := &TxInfo{
		receivedAmount: amountSatoshi,
		txOut:          outpoint,
	}

	chunk := &Chunk{
		addr:    bootstrapChunk.addr,
		status:  Retained,
		message: nil,
		txInfo:  txInfo,
	}

	wif, err := btcutil.DecodeWIF(bootstrapChunk.privKeyWif)
	if err != nil {
		log.Printf("error decoding privkey wif: %v", err)
		return nil, nil, err
	}

	return chunk, wif, nil
}

var (
	bootstrapChunks = []*BootstrapFeeChunk{
		&BootstrapFeeChunk{
			addr:       "mjadFfF2h3sNpU9iMETSiECCz7ArKdkx94",
			amount:     5.7,
			txId:       "a43ccc3b34ded6f18ccb5c066e421148a67334131f58067bbfc74e66a20d98b3",
			index:      0,
			privKeyWif: "92jwjgG3e7o8EcXAnYzjiVm3ukBawd34gQuYd7QuNaVUgLJc4Ue",
		},
	}
)

func BootstrapPool() {
	log.Printf("bootstrapping mix pool")
	bootstrapFeeChunks(bootstrapChunks)
}
