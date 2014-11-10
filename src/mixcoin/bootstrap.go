package mixcoin

import (
	"log"
	"reflect"

	"github.com/conformal/btcutil"
)

func (b *ReserveBootstrap) normalize() (*Utxo, *btcutil.WIF, error) {
	amount, err := btcutil.NewAmount(b.amount)

	utxo := &Utxo{
		Addr:   b.addr,
		Amount: amount,
		TxId:   b.txId,
		Index:  b.index,
	}

	wif, err := btcutil.DecodeWIF(b.wif)
	if err != nil {
		log.Printf("error decoding privkey wif: %v", err)
		return nil, nil, err
	}

	return utxo, wif, nil
}

type ReserveBootstrap struct {
	addr   string
	amount float64
	txId   string
	index  int
	wif    string
}

var (
	reserve = []*ReserveBootstrap{
		&ReserveBootstrap{
			addr:   "mjadFfF2h3sNpU9iMETSiECCz7ArKdkx94",
			amount: 5.6,
			txId:   "3ce5b0589bbcb592cf19d3d21243efcdabfbdd2beacd6d6c8f5a2a61ebea06ef",
			index:  1,
			wif:    "92jwjgG3e7o8EcXAnYzjiVm3ukBawd34gQuYd7QuNaVUgLJc4Ue",
		},
	}
)

func BootstrapPool() {
	var reserveAddresses []string
	for _, r := range reserve {
		reserveAddresses = append(reserveAddresses, r.addr)
	}
	log.Printf("bootstrapping mix pool with chunks: %v", reserveAddresses)
	for _, b := range reserve {
		utxo, wif, err := b.normalize()
		if err != nil {
			log.Printf("error parsing bootstrap reserve: %v", err)
		}
		err = rpc.ImportPrivKey(wif)
		if err != nil {
			log.Printf("error importing privkey: %v", err)
		}
		pool.Put(Reserve, utxo)
	}
}

func LoadReserves() {
	log.Printf("loading reserve utxos from db")
	items := db.Items()
	for _, item := range items {
		if isUtxo(item) {
			utxo := item.(*Utxo)
			log.Printf("read utxo from db: %+v", utxo)
		} else {
			msg := item.(*ChunkMessage)
			log.Printf("db has a leftover chunkmsg: %+v", msg)
			db.Delete(msg.Key())
		}
	}
}

func isUtxo(item PoolItem) bool {
	return reflect.TypeOf(item) == reflect.TypeOf(&Utxo{})
}
