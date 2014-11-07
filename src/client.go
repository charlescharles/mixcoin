package main

import (
	"./mixcoin"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"io/ioutil"
	"math/rand"
)

type MixcoinServer struct {
	Address string
	Name string
	PubKey string
}

func chooseServers(servers []*MixcoinServer, numMixes int) []*MixcoinServer {
	n := len(servers)
	var chosen []*MixcoinServer
	for int i := 0; i < numMixes; i++ {
		chosen = append(chosen, servers[rand.Intn(n)])
	}

	return chosen
} 

var (
	chunk = &mixcoin.ChunkMessage{
		Val:      4000000,
		SendBy:   306900,
		ReturnBy: 306950,
		OutAddr:  "charles",
		Fee:      2,
		Nonce:    123,
		Confirm:  1,
	}
}

func main() {
	marshaled, err := json.Marshal(chunk)
	if err != nil {
		panic(err)
	}

	reader := bytes.NewReader(marshaled)

	res, err := http.Post("http://localhost:8082/chunk", "application/json", reader)
	if err != nil {
		log.Printf("err")
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("error reading response body: %v", err)
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(bytes.NewBuffer(body))
	responseChunk := mixcoin.ChunkMessage{}
	decoder.Decode(&responseChunk)

	log.Printf("response: %v", responseChunk)
	log.Printf("mixaddr: %s", responseChunk.MixAddr)

	mixAddr := responseChunk.MixAddr

}

func verifyWarrant(msg *mixcoin.ChunkMessage, mixPubKey string) bool {

}
