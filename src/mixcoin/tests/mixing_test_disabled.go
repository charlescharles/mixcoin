package mixcoin

import (
	"github.com/conformal/btcjson"
	"github.com/conformal/btcrpcclient"
	"github.com/conformal/btcutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log"
	"testing"
)

func initializeMix(t *testing.T) *MockRpcClient {
	mockClient := new(MockRpcClient)

	newRpcClient = func(config *btcrpcclient.ConnConfig, ntfnHandlers *btcrpcclient.NotificationHandlers) RpcClient {
		log.Printf("creating mock rpc client")
		mockClient.newBlockHandler = ntfnHandlers.OnBlockConnected
		mockClient.recvTxHandler = ntfnHandlers.OnRecvTx
		return mockClient
	}

	mockClient.On("NotifyBlocks").Return(nil)
	mockClient.On("WalletPassphrase", mock.Anything, 7200).Return(nil)
	mockClient.On("ImportPrivKey", mock.Anything).Return(nil)

	log.Printf("starting server")

	StartMixcoinServer()

	return mockClient
}

func TestMix(t *testing.T) {
	mockClient := initializeMix(t)

	// send a chunk msg request
	mockClient.On("GetNewAddress").Return(mixAddr1, nil)
	mockClient.On("GetBestBlock").Return(blockHash1, int32(306328), nil)

	err := handleChunkRequest(chunkMsg1)

	assert.Nil(t, err)
	warrantLenGot := len(chunkMsg1.Warrant)
	warrantLenWant := len(exampleWarrant1)
	assert.Equal(t, warrantLenGot, warrantLenWant, "chunk msg should be signed")

	// block found
	unspentCall := mockClient.On("ListUnspentMinMaxAddresses", 0, 1, []btcutil.Address{mixAddr1})
	unspentCall.Return([]btcjson.ListUnspentResult{utxo1}, nil)

	mockClient.newBlockHandler(blockHash1, int32(306328))

	mockClient.AssertExpectations(t)
}
