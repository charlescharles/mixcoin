package mixcoin

import (
	"bytes"
	"fmt"
	"github.com/conformal/btcjson"
)

type Chunk struct {
	status  PoolType
	message *ChunkMessage
	txInfo  *TxInfo
	addr    string // the mix/escrow address
}

type ChunkMessage struct {
	Val      int64  `json:"val"`
	SendBy   int    `json:"sendBy"`
	ReturnBy int    `json:"returnBy"`
	OutAddr  string `json:"outAddr"`
	Fee      int    `json:"fee"`
	Nonce    int64  `json:"nonce"`
	Confirm  int    `json:"confirm"`

	MixAddr string `json:"mixAddr"`
	Warrant string `json:"warrant"`
}

type PoolManager interface {
	RegisterNewChunk(*ChunkMessage)
	RegisterReserveChunks([]*Chunk)
	RegisterReceived(string, *TxInfo, *btcwire.ShaHash)
	GetRandomChunk(PoolType) (*Chunk, error)
	GetReceivable() []btcutil.Address
	Prune()
}

// NOTE: assumes it's a single coin
func (chunk *Chunk) GetAsTxInput() (*btcjson.TransactionInput, error) {
	hash := chunk.txInfo.txOut.Hash
	index := chunk.txInfo.txOut.Index

	return &btcjson.TransactionInput{hash.String(), index}, nil
}

func (chunkMsg *ChunkMessage) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("\n-------------\n")

	buffer.WriteString(fmt.Sprintf(`
    val: %d
    sendby: %d
    returnby: %d
    outAddr: %sfee: %d
    nonce: %d
    confirm: %d
    mixAddr: %s
    warrant: %s`,
		chunkMsg.Val, chunkMsg.SendBy, chunkMsg.ReturnBy, chunkMsg.OutAddr, chunkMsg.Fee, chunkMsg.Nonce, chunkMsg.Confirm, chunkMsg.MixAddr, chunkMsg.Warrant))

	buffer.WriteString("\n-------------\n")

	return buffer.String()
}
