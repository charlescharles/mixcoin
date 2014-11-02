package mixcoin

import (
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestMix(t *testing.T) {
	var mockClient RpcClient

	newRpcClient = func(config *btcrpcclient.ConnConfig, ntfnHandlers *btcrpcclient.NotificationHandlers) RpcClient {
		log.Printf("creating mock rpc client")
		client := new(MockRpcClient)
		client.newBlockHandler = ntfnHandlers.OnBlockConnected
		client.recvTxHandler = ntfnHandlers.OnRecvTx
		mockClient = client
		return client
	}

	mockClient.Mock.On("NotifyBlocks").Return(nil)
	mockClient.Mock.On("WalletPassphrase", "Mixcoin1", 7200).Return(nil)

	StartMixcoinServer()

	mockClient.Mock.AssertExpectations(t)
}
