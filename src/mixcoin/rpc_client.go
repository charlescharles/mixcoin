package mixcoin

import (
	"btcrpcclient"
	"btcutil"
	"io/ioutil"
	"log"
)

var rpcClient *btcrpcclient.Client

func StartRpcClient() {
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
		log.Panicf("error creating rpc client")
	}

	// Register for block connect and disconnect notifications.
	if err = client.NotifyBlocks(); err != nil {
		log.Panicf("error setting notifyblock")
	}
}

func getNewAddress() (*btcutil.Address, error) {
	cfg := GetConfig()

	addr, err := rpcClient.GetNewAddress()
	if err != nil {
		rpcClient.CreateEncryptedWallet(cfg.WalletPass)
		addr, err = rpcClient.GetNewAddress()
	}
	if err != nil {
		log.Panicf("error getting new address")
		return nil, err
	}
	err = rpcClient.SetAccount(addr, cfg.MixAccount)
	if err != nil {
		log.Panicf("error setting account")
		return nil, err
	}
	return &addr, nil
}
