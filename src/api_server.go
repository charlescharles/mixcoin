package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	mixcoin "mixcoin/server"
	"net/http"
)

type ChunkRequest struct {
	val      int
	sendBy   int
	returnBy int
	outAdd   string
	fee      int
	nonce    int
	confirm  int
}

func handleChunkRequest(rw http.ResponseWriter, req *http.Request) {
	if err != nil {
		panic()
	}
	var chunkRequest ChunkRequest
	err = json.NewDecoder(req.Body).Decode(chunkRequest)

}

func main() {
	mixcoinServer := mixcoin.NewServer(opts)
	http.HandleFunc("/chunk", mixcoinServer.handleChunkRequest)
	log.Fatal(http.ListenAndServe(":8082", nil))
}
