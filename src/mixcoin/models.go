package mixcoin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/conformal/btcutil"
	"log"
)

type Utxo struct {
	Addr   string
	Amount btcutil.Amount
	TxId   string
	Index  int
}

func (u *Utxo) Key() string {
	return u.Addr
}

func (u *Utxo) Serialize() []byte {
	serialized, err := json.Marshal(*u)
	if err != nil {
		log.Panicf("error serializing utxo: %v", err)
	}

	return serialized
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

func (c *ChunkMessage) Serialize() []byte {
	serialized, err := json.Marshal(*c)
	if err != nil {
		log.Panicf("error serializing utxo: %v", err)
	}

	return serialized
}

func (chunkMsg *ChunkMessage) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("\n-------------\n")

	buffer.WriteString(fmt.Sprintf(`
    val: %d
    sendby: %d
    returnby: %d
    outAddr: %s
	fee: %d
    nonce: %d
    confirm: %d
    mixAddr: %s
    warrant: %s`,
		chunkMsg.Val, chunkMsg.SendBy, chunkMsg.ReturnBy, chunkMsg.OutAddr, chunkMsg.Fee, chunkMsg.Nonce, chunkMsg.Confirm, chunkMsg.MixAddr, chunkMsg.Warrant))

	buffer.WriteString("\n-------------\n")

	return buffer.String()
}

func validateChunkMsg(chunkMsg *ChunkMessage) error {
	if chunkMsg.Val != cfg.ChunkSize {
		return errors.New("Invalid chunk size")
	}
	if chunkMsg.Confirm < cfg.MinConfirmations {
		return errors.New("Invalid number of confirmations")
	}

	height := getBlockchainHeight()

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
	return nil
}
