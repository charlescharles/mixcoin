package main

import (
	"./mixcoin"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

var (
	chunk = &mixcoin.ChunkMessage{
		Val:      4000000,
		SendBy:   306890,
		ReturnBy: 306900,
		OutAddr:  "charles",
		Fee:      2,
		Nonce:    123,
		Confirm:  1,
	}
)

func main() {
	marshaled, err := json.Marshal(chunk)
	if err != nil {
		panic(err)
	}

	reader := bytes.NewReader(marshaled)

	res, err := http.Post("http://localhost:8082/chunk", "application/json", reader)
	log.Printf("response: %v", res.Body)
}
