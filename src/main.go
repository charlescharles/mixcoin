package main

import (
	"./mixcoin"
)

func main() {
	mixcoin.StartMixcoinServer()
	mixcoin.SendChunkTestnet()
	mixcoin.StartApiServer()
}
