package mixcoin

import (
	"btcutil"
)

func decodeAddress(encoded string) (btcutil.Address, error) {
	cfg := GetConfig()
	addr, err := btcutil.DecodeAddress(encoded, cfg.NetParams)

	if err != nil {
		return nil, err
	}

	return addr, nil
}
