package mixcoin

import (
	"btcnet"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func StartApiServer() {
	port := GetConfig().ApiPort
	http.HandleFunc("/chunk", apiHandleChunkRequest)
	log.Printf("listening on ", port)
	log.Fatal(http.ListenAndServe(strconv.Itoa(port), nil))
}

func apiHandleChunkRequest(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var chunk Chunk
	err := decoder.Decode(&chunk)
	if err != nil {
		log.Panicf("error decoding chunk: ", err)
	}

	chunkRes, err := handleChunkRequest(&chunk)
	if err != nil {
		log.Panicf("error handling chunk request: ", err)
		rw.WriteHeader(500)
		return
	}

	log.Println(chunkRes)
	json, err := json.Marshal(chunkRes)
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(500)
		return
	}
	_, err = rw.Write(json)
	if err != nil {
		panic(err)
	}
}
