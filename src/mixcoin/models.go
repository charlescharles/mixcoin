package mixcoin

import (
	"bytes"
	"fmt"
)

type Chunk struct {
	status  PoolType
	message *ChunkMessage
	txInfo  *TxInfo
}

type ChunkMessage struct {
	Val      int64  `json:"val"`
	SendBy   int    `json:"sendBy"`
	ReturnBy int    `json:"returnBy"`
	OutAddr  string `json:"outAddr"`
	Fee      int    `json:"fee"`
	Nonce    int    `json:"nonce"`
	Confirm  int    `json:"confirm"`

	MixAddr string `json:"mixAddr"`
	Warrant string `json:"warrant"`
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
