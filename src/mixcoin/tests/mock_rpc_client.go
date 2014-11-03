package mixcoin

import (
	"github.com/conformal/btcjson"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"github.com/conformal/btcws"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockRpcClient struct {
	mock.Mock
	newBlockHandler func(hash *btcwire.ShaHash, height int32)
	recvTxHandler   func(transaction *btcutil.Tx, details *btcws.BlockDetails)
}

func (o *MockRpcClient) NotifyBlocks() error {
	args := o.Mock.Called()
	return args.Error(0)
}

func (o *MockRpcClient) WalletPassphrase(pass string, timeout int64) error {
	args := o.Mock.Called(pass, timeout)
	return args.Error(0)
}

func (o *MockRpcClient) CreateEncryptedWallet(pass string) error {
	args := o.Mock.Called(pass)
	return args.Error(0)
}

func (o *MockRpcClient) GetNewAddress() (btcutil.Address, error) {
	args := o.Mock.Called()
	return args.Get(0).(btcutil.Address), args.Error(1)
}

func (o *MockRpcClient) GetBestBlock() (*btcwire.ShaHash, int32, error) {
	args := o.Mock.Called()
	return args.Get(0).(*btcwire.ShaHash), args.Get(1).(int32), args.Error(2)
}

func (o *MockRpcClient) CreateRawTransaction(inputs []btcjson.TransactionInput, amounts map[btcutil.Address]btcutil.Amount) (*btcwire.MsgTx, error) {
	args := o.Mock.Called(inputs, amounts)
	return args.Get(0).(*btcwire.MsgTx), args.Error(1)
}

func (o *MockRpcClient) SignRawTransaction(tx *btcwire.MsgTx) (*btcwire.MsgTx, bool, error) {
	args := o.Mock.Called(tx)
	return args.Get(0).(*btcwire.MsgTx), args.Bool(1), args.Error(2)
}

func (o *MockRpcClient) SendRawTransaction(tx *btcwire.MsgTx, allowHighFees bool) (*btcwire.ShaHash, error) {
	args := o.Mock.Called(tx, allowHighFees)
	return args.Get(0).(*btcwire.MsgTx), args.Error(1)
}

func (o *MockRpcClient) NotifyReceivedAsync(addrs []btcutil.Address) btcrpcclient.FutureNotifyReceivedResult {
	args := o.Mock.Called(addrs)
	return args.Get(0).(btcrpcclient.FutureNotifyReceivedResult)
}

func (o *MockRpcClient) ListUnspentMinMaxAddresses(min int, max int, addrs []btcutil.Address) ([]btcjson.ListUnspentResult, error) {
	args := o.Mock.Called(min, max, addrs)
	return args.Get(0).([]btcjson.ListUnspentResult), args.Error(1)
}

func mockNewRpcClient(config *btcrpcclient.ConnConfig, ntfnHandlers *btcrpcclient.NotificationHandlers) RpcClient {
	log.Printf("creating mock rpc client")
	client := new(MockRpcClient)
	client.newBlockHandler = ntfnHandlers.OnBlockConnected
	client.recvTxHandler = ntfnHandlers.OnRecvTx
	return client
}
