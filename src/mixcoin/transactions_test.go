package mixcoin

import (
	"github.com/conformal/btcjson"
	"github.com/conformal/btcnet"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"github.com/stretchr/testify/mock"
	"testing"
)

var (
	dest = "mzn21y12sj6fMEVghcwFHSuDjvtXo3HbU3"

	aliceUtxo = &Utxo{
		addr:   "mvA3EGSbpgGorow1oWzgAwrMEACmUnK4NU",
		amount: btcutil.Amount(300),
		txId:   "tx0",
		index:  0,
	}

	bobUtxo = &Utxo{
		addr:   "mxp4okhCVAddcViDXxop5rBXft6zPL9NNL",
		amount: btcutil.Amount(400),
		txId:   "tx1",
		index:  1,
	}

	aliceInput = btcjson.TransactionInput{
		Txid: "tx0",
		Vout: 0,
	}

	bobInput = btcjson.TransactionInput{
		Txid: "tx1",
		Vout: 1,
	}

	msgTx = btcwire.NewMsgTx()
)

func TestSend(t *testing.T) {
	rpc = NewMockRpcClient()
	pool = NewMockPool()
	cfg = GetConfig()

	cfg.TxFee = 70
	cfg.ChunkSize = 400
	cfg.NetParams = &btcnet.TestNet3Params

	pool.(*MockPool).On("Get", Reserve).Return(aliceUtxo, nil)
	pool.(*MockPool).On("Get", Mixing).Return(bobUtxo, nil)

	/*
		aliceAddr, _ := decodeAddress(aliceUtxo.addr)
		destAddr, _ := decodeAddress(dest)

			rpc.(*MockRpcClient).On("CreateRawTransaction",
				[]btcjson.TransactionInput{aliceInput, bobInput},
				map[btcutil.Address]btcutil.Amount{
					destAddr:  btcutil.Amount(400),
					aliceAddr: btcutil.Amount(293),
				}).Return(msgTx, nil)
	*/

	rpc.(*MockRpcClient).On("CreateRawTransaction",
		mock.AnythingOfType("[]btcjson.TransactionInput"),
		mock.AnythingOfType("map[btcutil.Address]btcutil.Amount")).Return(msgTx, nil)

	rpc.(*MockRpcClient).On("SignRawTransaction", msgTx).Return(msgTx, true, nil)

	txHash, _ := btcwire.NewShaHashFromStr("55b2bd88b2fa00626742cfb447bf483a49c6d5ca7ed8cb5033c64d3766dd6e11")
	rpc.(*MockRpcClient).On("SendRawTransaction", msgTx, mock.AnythingOfType("bool")).Return(txHash, nil)

	pool.(*MockPool).On("Put", Reserve, mock.AnythingOfType("*mixcoin.Utxo")).Return()

	send(dest)

	pool.(*MockPool).AssertExpectations(t)
	rpc.(*MockRpcClient).AssertExpectations(t)
}
