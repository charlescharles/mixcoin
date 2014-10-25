package mixcoin

import (
	"bytes"
	"fmt"
)

type Chunk struct {
	Val      int    `json:"val"`
	SendBy   int    `json:"sendBy"`
	ReturnBy int    `json:"returnBy"`
	OutAddr  string `json:"outAddr"`
	Fee      int    `json:"fee"`
	Nonce    int    `json:"nonce"`
	Confirm  int    `json:"confirm"`

	MixAddr string `json:"mixAddr"`
	Warrant string `json:"warrant"`

	TxInfo *TxInfo
}

func (chunk *Chunk) String() string {
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
		chunk.Val, chunk.SendBy, chunk.ReturnBy, chunk.OutAddr, chunk.Fee, chunk.Nonce, chunk.Confirm, chunk.MixAddr, chunk.Warrant))

	buffer.WriteString("\n-------------\n")

	return buffer.String()
}
