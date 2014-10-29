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
		OnBlockConnected: onNewBlock,
		//		OnRecvTx:         nil,
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
	}
	rpcClient = client
}

func getNewAddress() (*btcutil.Address, error) {
	cfg := GetConfig()
	addr, err := rpcClient.GetNewAddress()
	if err != nil {
		log.Printf("creating new encrypted wallet")
		rpcClient.CreateEncryptedWallet(cfg.WalletPass)
		addr, err = rpcClient.GetNewAddress()
	}
	if err != nil {
		log.Panicf("error getting new address")
		return nil, err
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
