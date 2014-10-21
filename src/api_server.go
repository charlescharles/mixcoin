package main

import (
	"btcnet"
	"encoding/json"
	"fmt"
	"log"
	"mixcoin"
	"net/http"
)

type ApiConfig struct {
	user          string
	pass          string
	walletAddress string
	apiPort       int
}

type ApiServer struct {
	config        *ApiConfig
	mixcoinServer *mixcoin.Server
}

func New(config *ApiConfig) (*ApiServer, error) {
	mixcoinConfig := &mixcoin.ServerConfig{
		RpcAddress:       config.walletAddress,
		RpcUser:          config.user,
		RpcPass:          config.pass,
		MinConfirmations: 6,
		ChunkSize:        2,
		PrivRingFile:     "/Users/cguo/.gnupg/secring.gpg",
		Passphrase:       "Thereis1",
		NetParams:        btcnet.SimNetParams,
	}

	mixcoinServer, err := mixcoin.NewServer(mixcoinConfig)

	if err != nil {
		panic("unable to create mixcoin server")
	}

	server := &ApiServer{
		config:        config,
		mixcoinServer: mixcoinServer,
	}

	return server, nil
}

func (self *ApiServer) Serve() {
	http.HandleFunc("/chunk", self.handleChunkRequest)
	fmt.Println("listening on ", self.config.apiPort)
	port := fmt.Sprintf(":%d", self.config.apiPort)
	log.Fatal(http.ListenAndServe(port, nil))
}

func (self *ApiServer) handleChunkRequest(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var chunk mixcoin.ChunkRequest
	err := decoder.Decode(&chunk)
	if err != nil {
		fmt.Println("err", err)
	}

	chunkRes, err := self.mixcoinServer.HandleChunkRequest(&chunk)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(chunkRes)
}

func main() {
	config := &ApiConfig{
		user:          "lightlike",
		pass:          "Thereisnone",
		walletAddress: "localhost:18554",
		apiPort:       8082,
	}
	server, err := New(config)
	if err != nil {
		panic("unable to create api server")
	}
	server.Serve()
}
