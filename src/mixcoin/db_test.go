package mixcoin

import (
	"fmt"
	"testing"

	"github.com/conformal/btcutil"
	"github.com/stretchr/testify/assert"
)

var (
	fakeChunkMsg = ChunkMessage{
		Val:      10,
		SendBy:   200,
		ReturnBy: 300,
		OutAddr:  "me",
		Fee:      3,
		Nonce:    123,
		Confirm:  1,
		MixAddr:  "charles",
	}

	fakeUtxo = Utxo{
		Addr:   "alice",
		Amount: btcutil.Amount(400),
		TxId:   "tx0",
		Index:  0,
	}
)

func TestSerialization(t *testing.T) {
	serialized := fakeChunkMsg.Serialize()
	msg := *deserialize(serialized).(*ChunkMessage)
	assert.Equal(t, fakeChunkMsg, msg, "ChunkMessage incorrectly de/serialized")

	serialized = fakeUtxo.Serialize()
	utxo := *deserialize(serialized).(*Utxo)
	assert.Equal(t, fakeUtxo, utxo, "UTXO incorrectly de/serialized")
}

func TestPersistence(t *testing.T) {
	testDbFile := "/Users/cguo/.mixcoin/db/testdb"
	db := NewMixcoinDB(testDbFile)

	var chunkAddrs []string
	chunkTable := make(map[string]*ChunkMessage)
	var utxoAddrs []string
	utxoTable := make(map[string]*Utxo)

	for i := 0; i < 10; i++ {
		chunkAddr := fmt.Sprintf("chunk%d", i)
		chunk := fakeChunkMsg
		chunk.MixAddr = chunkAddr

		utxoAddr := fmt.Sprintf("utxo%d", i)
		utxo := fakeUtxo
		utxo.Addr = utxoAddr

		chunkAddrs = append(chunkAddrs, chunkAddr)
		utxoAddrs = append(utxoAddrs, utxoAddr)
		chunkTable[chunkAddr] = &chunk
		utxoTable[utxoAddr] = &utxo
	}

	for _, chunkAddr := range chunkAddrs {
		db.Put(chunkTable[chunkAddr])
	}
	for _, utxoAddr := range utxoAddrs {
		db.Put(utxoTable[utxoAddr])
	}

	for _, chunkAddr := range chunkAddrs {
		want := chunkTable[chunkAddr]
		got := db.Get(chunkAddr)
		assert.Equal(t, want, got, "chunk incorrectly retrieved")
		db.Delete(chunkAddr)
	}

	for _, utxoAddr := range utxoAddrs {
		want := utxoTable[utxoAddr]
		got := db.Get(utxoAddr)
		assert.Equal(t, want, got, "utxo incorrectly retrieved")
		db.Delete(utxoAddr)
	}

	items := db.Items()
	assert.Equal(t, len(items), 0, fmt.Sprintf("db still contains %d items: %+v", len(items), items))
}
