package mixcoin

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/conformal/btcutil"
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
	return fmt.Sprintf(`
	{
		val: \t%d
		sendby: \t%d
		returnby: \t%d
		outAddr: \t%s
		fee: \t%d
		nonce: \t%d
		confirm: \t%d
		mixAddr: \t%s
		warrant: \t%s
	}`,
		chunkMsg.Val, chunkMsg.SendBy, chunkMsg.ReturnBy, chunkMsg.OutAddr, chunkMsg.Fee, chunkMsg.Nonce, chunkMsg.Confirm, chunkMsg.MixAddr, chunkMsg.Warrant)
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
