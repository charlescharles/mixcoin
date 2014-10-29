package mixcoin

import (
	"btcnet"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	RpcAddress string // host:port for btcwallet instance
	RpcUser    string // username for btcwallet instance
	RpcPass    string // password for btcwallet instance
	CertFile   string // path to server cert file
	MixAccount string // name for account containing mix addresses
	WalletPass string // wallet passphrase

	NetParams *btcnet.Params // network type: simnet, mainnet, etc.
	ApiPort   int            // port to listen on for /chunk requests

	MinConfirmations int   // min confirmations we require
	ChunkSize        int64 // standard chunk size, satoshis

	PrivRingFile string // path to privring
	Passphrase   string // password for privring
}

func GetConfig() *Config {
	return &config
}

var defaultConfig = Config{
	RpcAddress: "127.0.0.1:18554",
	RpcUser:    "mixcoin",
	RpcPass:    "Mixcoin1",
	CertFile:   os.Getenv("HOME") + "/.mixcoin/mixcoinCA.cer",
	MixAccount: "mixcoin",
	WalletPass: "Mixcoin1",

	NetParams: &btcnet.SimNetParams,
	ApiPort:   8082,

	MinConfirmations: 6,
	ChunkSize:        1000,

	PrivRingFile: os.Getenv("HOME") + "/.mixcoin/secring.gpg",
	Passphrase:   "Thereis1",
}

var config Config

func init() {
	configFile := os.Getenv("HOME") + "/.mixcoin/config.json"
	log.Printf("Reading " + configFile)

	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		writeDefaultConfig(configFile)
		fmt.Println("Config file written to ~/.mixcoin/config.json. Please edit and re-run.")
	}

	config = Config{}
	configBuf := bytes.NewBuffer(configBytes)
	//err = json.Unmarshal(configBuf, &config)
	decoder := json.NewDecoder(configBuf)
	err = decoder.Decode(&config)
	if err != nil {
		log.Panicf("Invalid configuration file %s: %v", configFile, err)
	}
}

func writeDefaultConfig(configFile string) {
	log.Printf("Creating default config file %s", configFile)
	configBytes, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(filepath.Dir(configFile), 0700)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(configFile, configBytes, 0600)
	if err != nil {
		panic(err)
	}
}
