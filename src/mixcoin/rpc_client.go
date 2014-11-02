package mixcoin

import (
	"btcrpcclient"
	"btcutil"
	"io/ioutil"
	"log"
)

var rpcClient *btcrpcclient.Client

func StartRpcClient() {
	log.Println("starting rpc client")

	cfg := GetConfig()

	log.Printf("Reading btcd cert file %s", cfg.CertFile)
	certs, err := ioutil.ReadFile(cfg.CertFile)
	if err != nil {
		log.Panicf("couldn't read btcd certs")
	}

	connCfg := &btcrpcclient.ConnConfig{
		Host:         cfg.RpcAddress,
		Endpoint:     "ws",
		User:         cfg.RpcUser,
		Pass:         cfg.RpcPass,
		Certificates: certs,
		DisableTLS:   false,
	}

	ntfnHandlers := btcrpcclient.NotificationHandlers{
		OnBlockConnected: onBlockConnected,
		OnRecvTx:         onRecvTx,
	}

	client, err := btcrpcclient.New(connCfg, &ntfnHandlers)
	if err != nil {
		log.Panicf("error creating rpc client: %v", err)
	}

	// Register for block connect and disconnect notifications.
	if err = client.NotifyBlocks(); err != nil {
		log.Panicf("error setting notifyblock")
	}

	log.Printf("unlocking wallet")

	err = client.WalletPassphrase(cfg.WalletPass, 7200)
	if err != nil {
		log.Printf("error unlocking wallet: %v", err)
		log.Printf("trying to create a new wallet:")
		err = client.CreateEncryptedWallet(cfg.WalletPass)
		if err != nil {
			log.Panicf("error creating new wallet: %v", err)
		}
		log.Printf("successfully created a new wallet")
	}
	rpcClient = client
}

func getNewAddress() (*btcutil.Address, error) {
	addr, err := rpcClient.GetNewAddress()
	if err != nil {
		log.Panicf("error getting new address: %v", err)
	}
	/**
	err = rpcClient.SetAccount(addr, cfg.MixAccount)
	if err != nil {
		log.Panicf("error setting account: %v", err)
		return nil, err
	}
	*/
	return &addr, nil
}

// TODO only update occasionally; no need to check every time
func getBlockchainHeight() (int, error) {
	log.Printf("getting blockchain height")
	_, height32, err := rpcClient.GetBestBlock()
	if err != nil {
		return -1, err
	}
	log.Printf("got blockchain height: %v", height32)
	return int(height32), nil
}
