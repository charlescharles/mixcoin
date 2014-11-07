package mixcoin

import (
	"github.com/conformal/btcutil"
	"log"
)

func (b *ReserveBootstrap) normalize() (*Utxo, *btcutil.WIF, error) {
	amount, err := btcutil.NewAmount(b.amount)

	utxo := &Utxo{
		addr:   b.addr,
		amount: amount,
		txId:   b.txId,
		index:  b.index,
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
			amount: 5.7,
			txId:   "4a708f8563a074b47585f0830d75b8afb3c8073fda2b972e2388f50a2eb03bc4",
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
