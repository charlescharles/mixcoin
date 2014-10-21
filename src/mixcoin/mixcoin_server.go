package mixcoin

import (
	"btcnet"
	"btcrpcclient"
	"btcutil"
	"btcwire"
	"btcws"

	"bytes"
	"code.google.com/p/go.crypto/openpgp"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type chunk struct {
	val      int
	sendBy   int
	returnBy int
	outAddr  *btcutil.AddressPubKeyHash
	fee      int
	nonce    int
	confirm  int
}

type chunkTable struct {
	receivable map[string]*chunk
	pool       map[string]*chunk
	retained   map[string]*chunk
}

type ServerConfig struct {
	// host:port for btcwallet instance
	RpcAddress string
	// username for btcwallet instance
	RpcUser string
	// password for btcwallet instance
	RpcPass string

	// min confirmations we require
	MinConfirmations int
	// standard chunk size
	ChunkSize int

	// path to pubring
	PubRingFile string
	// path to privring
	PrivRingFile string
	// password for privring
	Passphrase string

	NetParams *btcnet.Params
}

func NewChunkTable() *chunkTable {
	table := &chunkTable{
		receivable: make(map[string]*chunk),
		pool:       make(map[string]*chunk),
		retained:   make(map[string]*chunk),
	}
	return table
}

type Server struct {
	config *ServerConfig
	chunks *chunkTable
	rpc    *btcrpcclient.Client
}

func NewServer(config *ServerConfig) (*Server, error) {

	//btcdHomeDir := btcutil.AppDataDir("btcd", false)
	certs, err := ioutil.ReadFile("/Users/cguo/server.crt")
	//certs, err := ioutil.ReadFile(filepath.Join(btcdHomeDir, "rpc.cert"))
	if err != nil {
		panic("couldn't read btcd certs")
	}
	connCfg := &btcrpcclient.ConnConfig{
		Host:         config.RpcAddress,
		Endpoint:     "ws",
		User:         config.RpcUser,
		Pass:         config.RpcPass,
		Certificates: certs,
		DisableTLS:   false,
	}

	client, err := btcrpcclient.New(connCfg, &ntfnHandlers)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Register for block connect and disconnect notifications.
	if err = client.NotifyBlocks(); err != nil {
		fmt.Println(err)
		return nil, err
	}

	server := &Server{
		config: config,
		chunks: NewChunkTable(),
		rpc:    client,
	}

	ntfnHandlers := btcrpcclient.NotificationHandlers{
		OnBlockConnected: server.onNewBlock,
		OnBlockDisconnected: func(hash *btcwire.ShaHash, height int32) {
			log.Printf("Block disconnected: %v (%d)", hash, height)
		},
	}

	return server, nil
}

func (self *Server) HandleChunkRequest(chunkReq *ChunkRequest) (*ChunkRequest, error) {
	fmt.Println("handling chunk request")

	escrowAddr, err := self.getNewAddress()
	if err != nil {
		fmt.Println(err)
		panic("unable to create new address")
	}
	encodedAddr := (*escrowAddr).EncodeAddress()

	chunkReq.EscrowAddr = encodedAddr

	err = self.signChunk(chunkReq)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	self.registerNewChunk(encodedAddr, chunkReq)
	self.registerAddress(encodedAddr)

	return chunkReq, nil
}

func (self *Server) getNewAddress() (*btcutil.Address, error) {
	addr, err := self.rpc.GetNewAddress()
	if err != nil {
		self.rpc.CreateEncryptedWallet("Thereis1")
	}
	addr, err = self.rpc.GetNewAddress()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &addr, nil
}

func (self *Server) signChunk(chunk *ChunkRequest) error {
	fmt.Println("signing chunk")
	marshaledBytes, _ := json.Marshal(chunk)
	marshaledBuf := bytes.NewBuffer(marshaledBytes)

	keyringFileBuffer, err := os.Open(self.config.PrivRingFile)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer keyringFileBuffer.Close()
	entityList, err := openpgp.ReadKeyRing(keyringFileBuffer)
	if err != nil {
		fmt.Println(err)
		return err
	}
	entity := entityList[0]
	passphrasebyte := []byte(self.config.Passphrase)
	entity.PrivateKey.Decrypt(passphrasebyte)
	for _, subkey := range entity.Subkeys {
		subkey.PrivateKey.Decrypt(passphrasebyte)
	}
	armoredSigBuf := new(bytes.Buffer)
	err = openpgp.ArmoredDetachSign(armoredSigBuf, entity, marshaledBuf, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}
	armoredSigEnc, err := ioutil.ReadAll(armoredSigBuf)
	if err != nil {
		fmt.Println(err)
		return err
	}
	armoredSig := base64.StdEncoding.EncodeToString(armoredSigEnc)

	chunk.Warrant = armoredSig

	return nil
}

/*
* TODO: ChunkRequest -> Chunk
 */
func (self *Server) registerNewChunk(encodedAddr string, chunk *ChunkRequest) error {
	return nil
}

func (self *Server) registerAddress(encodedAddr string) error {
	addr, err := btcutil.DecodeAddress(encodedAddr, self.config.NetParams)
	self.rpc.NotifyReceived([]*btcutil.Address{addr})
}

func (self *Server) onNewBlock(hash *btcwire.ShaHash, height int32) {
	minConf := self.config.MinConfirmations
	stdChunkSize := self.config.ChunkSize

	for encodedAddr, chunk := range self.chunks.receivable {
		addr := btcutil.DecodeAddress(encod, self.config.NetParams)
		amount, err := self.rpc.GetReceivedByAddressMinConf(addr, minConf)
		// TODO check that the time is before receivedBy
		if amount >= stdChunkSize {
			// check random beacon to see if we should retain as fee
			// move chunk from receivable into pool
		}
	}
}
