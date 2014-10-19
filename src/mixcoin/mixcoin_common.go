package mixcoin

import (
	"bytes"
	"fmt"
)

type ChunkRequest struct {
	Val      int
	SendBy   int
	ReturnBy int
	OutAddr  string
	Fee      int
	Nonce    int
	Confirm  int

	EscrowAddr string
	Warrant    string
}

func (chunk *ChunkRequest) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("\n-------------\n")

	buffer.WriteString(fmt.Sprintf(`
    val: %d
    sendby: %d
    returnby: %d
    outAddr: %sfee: %d
    nonce: %d
    confirm: %d
    escrowAddr: %s
    warrant: %s`,
		chunk.Val, chunk.SendBy, chunk.ReturnBy, chunk.OutAddr, chunk.Fee, chunk.Nonce, chunk.Confirm, chunk.EscrowAddr, chunk.Warrant))

	buffer.WriteString("\n-------------\n")

	return buffer.String()
}
