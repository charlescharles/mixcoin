package mixcoin

import (
	"github.com/conformal/btcutil"
)

func decodeAddress(encoded string) (btcutil.Address, error) {
	addr, err := btcutil.DecodeAddress(encoded, cfg.NetParams)

	if err != nil {
		return nil, err
	}

	return addr, nil
}
