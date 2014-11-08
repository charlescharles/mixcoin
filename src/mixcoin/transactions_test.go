package mixcoin

import (
	"github.com/conformal/btcjson"
	"github.com/conformal/btcnet"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"github.com/stretchr/testify/mock"
	"log"
	"testing"
)

var (
	dest = "mzn21y12sj6fMEVghcwFHSuDjvtXo3HbU3"

	aliceUtxo = &Utxo{
		Addr:   "mvA3EGSbpgGorow1oWzgAwrMEACmUnK4NU",
		Amount: btcutil.Amount(300),
		TxId:   "tx0",
		Index:  0,
	}

	bobUtxo = &Utxo{
		Addr:   "mxp4okhCVAddcViDXxop5rBXft6zPL9NNL",
		Amount: btcutil.Amount(400),
		TxId:   "tx1",
		Index:  1,
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

	aliceAddr, _ := decodeAddress(aliceUtxo.Addr)
	destAddr, _ := decodeAddress(dest)

	var expectedTxInputs []btcjson.TransactionInput
	expectedTxInputs = append(expectedTxInputs, aliceInput)
	expectedTxInputs = append(expectedTxInputs, bobInput)

	expectedOutMap := make(map[btcutil.Address]btcutil.Amount)
	expectedOutMap[aliceAddr] = btcutil.Amount(400)
	expectedOutMap[destAddr] = btcutil.Amount(293)

	/*
		rpc.(*MockRpcClient).On("CreateRawTransaction",
			mock.AnythingOfType("[]btcjson.TransactionInput"),
			mock.AnythingOfType("map[btcutil.Address]btcutil.Amount")).Return(msgTx, nil)
	*/

	log.Printf("expectedTxInputs: %v", expectedTxInputs)
	log.Printf("expectedOutMap: %v", expectedOutMap)
	rpc.(*MockRpcClient).On("CreateRawTransaction", expectedTxInputs, expectedOutMap).Return(msgTx, nil)

	rpc.(*MockRpcClient).On("SignRawTransaction", msgTx).Return(msgTx, true, nil)

	txHash, _ := btcwire.NewShaHashFromStr("55b2bd88b2fa00626742cfb447bf483a49c6d5ca7ed8cb5033c64d3766dd6e11")
	rpc.(*MockRpcClient).On("SendRawTransaction", msgTx, mock.AnythingOfType("bool")).Return(txHash, nil)

	pool.(*MockPool).On("Put", Reserve, mock.AnythingOfType("*mixcoin.Utxo")).Return()

	send(dest)

	pool.(*MockPool).AssertExpectations(t)
	rpc.(*MockRpcClient).AssertExpectations(t)
}
