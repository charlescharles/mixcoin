package mixcoin

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func StartApiServer() {
	port := cfg.ApiPort
	http.HandleFunc("/chunk", apiHandleChunkRequest)
	log.Printf("listening on %v", port)
	portStr := ":" + strconv.Itoa(port)
	log.Fatal(http.ListenAndServe(portStr, nil))
}

func apiHandleChunkRequest(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var chunkMsg ChunkMessage
	err := decoder.Decode(&chunkMsg)
	if err != nil {
		log.Panicf("error decoding chunk: ", err)
		return
	}

	err = handleChunkRequest(&chunkMsg)
	if err != nil {
		log.Printf("error handling chunk request: ", err)
		rw.WriteHeader(500)
		return
	}

	json, err := json.Marshal(chunkMsg)
	if err != nil {
		log.Printf("error marshaling json: %v", err)
		rw.WriteHeader(500)
		return
	}
	_, err = rw.Write(json)
	if err != nil {
		panic(err)
	}
}
