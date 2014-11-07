package main

import (
	"./mixcoin"
)

func main() {
	mixcoin.StartMixcoinServer()
	mixcoin.StartApiServer()
}
