package mixcoin

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/conformal/btcnet"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	RpcAddress string // host:port for btcwallet instance
	RpcUser    string // username for btcwallet instance
	RpcPass    string // password for btcwallet instance
	CertFile   string // path to server cert file
	MixAccount string // name for account containing mix addresses
	WalletPass string // wallet passphrase

	NetParamName string         // the string indicating which net
	NetParams    *btcnet.Params // network type: simnet, mainnet, etc.
	ApiPort      int            // port to listen on for /chunk requests

	MinConfirmations   int    // min confirmations we require
	ChunkSize          int64  // standard chunk size, satoshis
	MaxFutureChunkTime int    // max block height delta into the future to accept chunk contracts
	TxFee              int64  // standard miner fee, satoshis
	DbFile             string // path to chunk database (for server)

	PrivRingFile string // path to privring
	Passphrase   string // password for privring

}

var defaultConfig = Config{
	RpcAddress: "127.0.0.1:18332",
	RpcUser:    "mixcoin",
	RpcPass:    "Mixcoin1",
	CertFile:   os.Getenv("HOME") + "/.mixcoin/mixcoinCA.cer",
	MixAccount: "mixcoin",
	WalletPass: "Mixcoin1",

	NetParamName: "testnet",
	ApiPort:      8082,

	MinConfirmations:   1,
	ChunkSize:          4000000,
	MaxFutureChunkTime: 72,    // a bit less than 12 hours
	TxFee:              10000, // 10k satoshis
	DbFile:             os.Getenv("HOME") + "/.mixcoin/db/mixcoin.db",

	PrivRingFile: os.Getenv("HOME") + "/.mixcoin/secring.gpg",
	Passphrase:   "Thereis1",
}

func validateConfig(cfg *Config) error {
	if cfg.RpcAddress == "" {
		return errors.New("RpcAddress must be set")
	}
	if cfg.RpcUser == "" {
		return errors.New("RpcUser must be set")
	}
	if cfg.RpcPass == "" {
		return errors.New("RpcPass must be set")
	}
	if cfg.CertFile == "" {
		return errors.New("CertFile must be set")
	}
	if cfg.MixAccount == "" {
		return errors.New("MixAccount must be set")
	}
	if cfg.WalletPass == "" {
		return errors.New("WalletPass must be set")
	}
	if cfg.NetParamName == "" {
		return errors.New("NetParamName must be set")
	}
	if cfg.ApiPort == 0 {
		return errors.New("ApiPort must be set")
	}
	if cfg.DbFile == "" {
		return errors.New("DbFile must be set")
	}
	if cfg.PrivRingFile == "" {
		return errors.New("PrivRingFile must be set")
	}
	if cfg.Passphrase == "" {
		return errors.New("Passphrase must be set")
	}
	if cfg.MinConfirmations < 0 {
		return errors.New("MinConfirmations must be >= 0")
	}
	if cfg.ChunkSize == 0 {
		return errors.New("ChunkSize must be > 0")
	}
	if cfg.TxFee < 0 {
		return errors.New("TxFee must be >= 0")
	}
	if _, err := url.Parse(cfg.RpcAddress); err != nil {
		return errors.New("invalid RPC address")
	}

	return nil
}

func GetConfig() *Config {
	configFile := os.Getenv("HOME") + "/.mixcoin/config.json"
	log.Printf("Reading " + configFile)

	// read bytes, write file and exit if not present
	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		writeDefaultConfig(configFile)
		log.Println("Config file written to ~/.mixcoin/config.json. Please edit and re-run.")
		os.Exit(1)
		return nil
	}

	// unmarshal into config
	config := Config{}
	configBuf := bytes.NewBuffer(configBytes)
	decoder := json.NewDecoder(configBuf)
	if err = decoder.Decode(&config); err != nil {
		log.Panicf("Invalid configuration file %s: %v", configFile, err)
	}

	if err = validateConfig(&config); err != nil {
		log.Printf("Invalid configuration file: %v", err)
		os.Exit(1)
		return nil
	}

	// set netparams
	if err = parseConfig(&config); err != nil {
		log.Panicf("Invalid configuration file: %v", err)
	}

	return &config
}

func parseConfig(config *Config) error {
	switch strings.ToLower(config.NetParamName) {
	case "testnet":
		config.NetParams = &btcnet.TestNet3Params
	case "mainnet":
		config.NetParams = &btcnet.MainNetParams
	case "simnet":
		config.NetParams = &btcnet.SimNetParams
	default:
		return errors.New("unrecognized net param name")
	}
	return nil
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
