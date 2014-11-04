package mixcoin

import (
	"bytes"
	"fmt"
	"github.com/conformal/btcjson"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
)

type Utxo struct {
	addr   string
	amount btcutil.Amount
	txId   string
	index  int
}

func (u *Utxo) Key() string {
	return u.addr
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

func (c *ChunkMessage) Key() string {
	return c.MixAddr
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

func validateChunkMsg(chunkMsg *ChunkMessage) error {
	cfg := GetConfig()

	if chunkMsg.Val != cfg.ChunkSize {
		return errors.New("Invalid chunk size")
	}
	if chunkMsg.Confirm < cfg.MinConfirmations {
		return errors.New("Invalid number of confirmations")
	}

	height, err := getBlockchainHeight()
	if err != nil {
		return err
	}
	blockchainHeight = height

	if chunkMsg.SendBy-blockchainHeight > cfg.MaxFutureChunkTime {
		return errors.New("sendby time too far in the future")
	}
	if chunkMsg.SendBy <= blockchainHeight {
		return errors.New("sendby time has already passed")
	}
	if chunkMsg.ReturnBy-chunkMsg.SendBy < 2 {
		return errors.New("not enough time between sendby and returnby")
	}
	log.Printf("validated block")
	return nil
}
