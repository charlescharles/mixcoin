package mixcoin

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
)

type DB interface {
	Put(PoolItem)
	Get(string) PoolItem
	Delete(string)
	Close()
	Items() map[string]PoolItem
}

type MixcoinDB struct {
	db *leveldb.DB
}

func NewMixcoinDB(path string) DB {
	db, err := leveldb.OpenFile(path, nil)

	if err != nil {
		log.Panicf("couldn't create or open db file at: %s", cfg.DbFile)
	}

	return &MixcoinDB{db}
}

func (m *MixcoinDB) Put(item PoolItem) {
	key := []byte(item.Key())
	val := item.Serialize()
	err := m.db.Put(key, val, nil)
	if err != nil {
		log.Panicf("error persisting item: %v", err)
	}
}

func (m *MixcoinDB) Get(key string) PoolItem {
	k := []byte(key)
	serialized, err := m.db.Get(k, nil)
	if err != nil {
		log.Panicf("error retrieving key %s from db: %v", key, err)
	}
	return deserialize(serialized)

}

func (m *MixcoinDB) Delete(key string) {
	k := []byte(key)
	err := m.db.Delete(k, nil)
	if err != nil {
		log.Panicf("error deleting key %s from db: %v", key, err)
	}
}

func (m *MixcoinDB) Close() {
	err := m.db.Close()
	if err != nil {
		log.Panicf("error closing db: %v", err)
	}
}

func (m *MixcoinDB) Items() map[string]PoolItem {
	ret := make(map[string]PoolItem)
	iter := m.db.NewIterator(nil, nil)
	for iter.Next() {
		key := string(iter.Key())
		val := iter.Value()
		item := deserialize(val)
		ret[key] = item
	}
	return ret
}

func deserialize(item []byte) PoolItem {
	// first try deserializing as a chunkmsg
	chunkMsg := ChunkMessage{}
	err := json.Unmarshal(item, &chunkMsg)
	if (err == nil) && (chunkMsg != ChunkMessage{}) {
		return &chunkMsg
	}

	utxo := Utxo{}
	err = json.Unmarshal(item, &utxo)
	if (err == nil) && (utxo != Utxo{}) {
		return &utxo
	}
	log.Printf("item: %s", string(item))
	log.Panicf("couldn't deserialize db item: %v", err)
	return nil
}
