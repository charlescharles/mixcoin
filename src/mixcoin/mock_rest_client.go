package mixcoin

import (
	"github.com/conformal/btcjson"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
)

var (
	chunkMsg1 = &ChunkMessage{
		Val:      4000000,
		SendBy:   306330,
		ReturnBy: 306340,
		OutAddr:  "muVYg4SC4fBr5xHR9K2Ym7nb2WcMLSWXPG",
		Fee:      2,
		Nonce:    2909333,
		Confirm:  1,
	}

	exampleWarrant1 = "LS0tLS1CRUdJTiBQR1AgU0lHTkFUVVJFLS0tLS0KCndzQmNCQUFCQ0FBUUJRSlVXQkNrQ1JBM0VaWFY3OW1yRVFBQUsxTUlBTDR3SmhDS2M2dU5LV3F4aG9VdFVHZ3kKcEtXZ3ZqOUZiL3FKU01vYmQ1QVFHaklsTHNMR2hxYXpEUFdxN0VuRXZPRlkvMEljSGUxby9LdUszemZXcnlYKwowd0tnMkE2enNuMjc2MFk2eDN3bFdBOUFxSElaaFBieG42RS92dEphSkdZTzdyVUdXRmsxdFpIMFRUTi93aytlCmh0TnpZTjY0Q2JMMkR0TTI5SnNFR0s2NnltbFVsTVhzZU4wczRSSWEybmV1NzdQYm5oV3NSTEcxYVpkZ0VFUS8KYSt4T1pSOTVpWGFvK3M4NU9XU3JBc2dJRG9ibnNVdUlva0hPMmlReFVEbG83ZGtMK2w1QkM0MHVsYUVlTEd0dgpjWkhsSGtTZlFCY0hYRUZCNURwTGZyTkt1cXFRMDkrMHkvd081cVVaMSsySVpvYkNhUUh0Q0VoYjE4K2JXSVk9Cj1BNXJiCi0tLS0tRU5EIFBHUCBTSUdOQVRVUkUtLS0tLQ=="
)

var (
	mixAddr1   btcutil.Address
	blockHash1 *btcwire.ShaHash
	utxo1      btcjson.ListUnspentResult
)

func init() {
	mixAddr1, _ = decodeAddress("mzm9kTwcChxuVMZnbaook3qSwVixixgwMz")
	blockHash1, _ = btcwire.NewShaHashFromStr("123123123")
	utxo1 = btcjson.ListUnspentResult{
		TxId:          "tx0",
		Vout:          uint32(0),
		Amount:        4000000,
		Confirmations: 1,
		Address:       mixAddr1.EncodeAddress(),
	}
}
