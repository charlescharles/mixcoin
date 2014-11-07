package main

import (
	"./mixcoin"
	"github.com/conformal/btcnet"
	"github.com/conformal/btcrpcclient"
	"github.com/conformal/btcutil"
	"log"
)

var (
	rpc mixcoin.RpcClient
)

func sendTo(dest string, amt int) {
	addr, err := decodeAddress(dest)
	if err != nil {
		panic(err)
	}
	hash, err := rpc.SendToAddress(addr, btcutil.Amount(amt))

	if err != nil {
		log.Printf("error sending tx: %v", err)
	}
	log.Printf("sent tx with hash %v", hash)
}

func decodeAddress(addr string) (btcutil.Address, error) {
	return btcutil.DecodeAddress(addr, &btcnet.TestNet3Params)
}

func main() {
	cfg = GetConfig()
	rpc = mixcoin.NewRpcClient().(*btcrpcclient.Client)

	addr := "mtbrGQFXpEDXq9yVdNEC3JeUbvJQcj5Jye"
	amt := 4000000
	sendTo(addr, amt)
}
