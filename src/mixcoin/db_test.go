package mixcoin

import (
	"fmt"
	"github.com/conformal/btcutil"
	"reflect"
	"testing"
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
	msg := deserialize(serialized).(*ChunkMessage)
	if !reflect.DeepEqual(fakeChunkMsg, *msg) {
		err := "chunk msg incorrectly deserialized:\n"
		err += fmt.Sprintf("want: %v\ngot: %v\n", fakeChunkMsg, *msg)
		t.Errorf(err)
	}

	serialized = fakeUtxo.Serialize()
	utxo := deserialize(serialized).(*Utxo)
	if !reflect.DeepEqual(fakeUtxo, *utxo) {
		err := "utxo incorrectly deserialized:\n"
		err += fmt.Sprintf("want: %v\ngot: %v\n", fakeUtxo, *utxo)
		t.Errorf(err)
	}
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
		if !reflect.DeepEqual(want, got) {
			t.Errorf("want chunk: %v\ngot chunk: %v\n", want, got)
		}
		db.Delete(chunkAddr)
	}

	for _, utxoAddr := range utxoAddrs {
		want := utxoTable[utxoAddr]
		got := db.Get(utxoAddr)
		if !reflect.DeepEqual(want, got) {
			t.Errorf("want utxo: %v\ngot utxo: %v\n", want, got)
		}
		db.Delete(utxoAddr)
	}

	items := db.Items()
	if len(items) != 0 {
		t.Errorf("db still contains %d items: %v", len(items), items)
	}
}
