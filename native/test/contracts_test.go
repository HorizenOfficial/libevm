package test

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"libevm/interop"
	"libevm/lib"
	"math/big"
	"testing"
)

func call[T any](t *testing.T, instance *lib.Service, method string, args interface{}) T {
	jsonArgs := ""
	if args != nil {
		jsonBytes, err := json.Marshal(args)
		if err != nil {
			panic(err)
		}
		jsonArgs = string(jsonBytes)
	}
	//t.Log("invoke", method, jsonArgs)
	result := interop.Invoke(instance, method, jsonArgs)
	var wrappedResponse struct {
		Error  string `json:"error"`
		Result T      `json:"result"`
	}
	err := interop.Deserialize(result, &wrappedResponse)
	if err != nil {
		t.Fatalf("invalid invocation result: %v", err)
	}
	if wrappedResponse.Error != "" {
		t.Errorf("invocation failed: %v", err)
	}
	//t.Log("response", toJsonResponse(err, result))
	return wrappedResponse.Result
}

// integration test using Solidity contracts and the Invoke interface
func TestContracts(t *testing.T) {
	var (
		instance     = lib.New()
		user         = common.HexToAddress("0xbafe3b6f2a19658df3cb5efca158c93272ff5c0b")
		emptyHash    = common.HexToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
		initialValue = common.Big0
		anotherValue = big.NewInt(5555)
	)

	dbHandle := call[int](t, instance, "OpenLevelDB", lib.LevelDBParams{Path: t.TempDir()})
	handle := call[int](t, instance, "StateOpen", lib.StateParams{
		DatabaseParams: lib.DatabaseParams{DatabaseHandle: dbHandle},
		Root:           emptyHash,
	})
	call[any](t, instance, "StateAddBalance", lib.BalanceParams{
		AccountParams: lib.AccountParams{
			HandleParams: lib.HandleParams{Handle: handle},
			Address:      user,
		},
		Amount: (*hexutil.Big)(big.NewInt(1000000000000000000)),
	})
	call[any](t, instance, "StateSetNonce", lib.NonceParams{
		AccountParams: lib.AccountParams{
			HandleParams: lib.HandleParams{Handle: handle},
			Address:      user,
		},
		Nonce: 1,
	})
	// deploy contract
	result := call[lib.EvmResult](t, instance, "EvmApply", lib.EvmParams{
		HandleParams: lib.HandleParams{Handle: handle},
		From:         user,
		To:           nil,
		Input:        Storage.Deploy(initialValue),
		AvailableGas: 200000,
		GasPrice:     (*hexutil.Big)(big.NewInt(1000000000)),
	})
	if result.EvmError != "" {
		t.Fatalf("vm error: %v", result.EvmError)
	}
	getCodeResult := call[[]byte](t, instance, "StateGetCode", lib.AccountParams{
		HandleParams: lib.HandleParams{Handle: handle},
		Address:      *result.ContractAddress,
	})
	if common.Bytes2Hex(Storage.RuntimeCode()) != common.Bytes2Hex(getCodeResult) {
		t.Fatalf("deployed code does not match %s", common.Bytes2Hex(getCodeResult))
	}
	// call function to store value
	call[any](t, instance, "EvmApply", lib.EvmParams{
		HandleParams: lib.HandleParams{Handle: handle},
		From:         user,
		To:           result.ContractAddress,
		Input:        Storage.Store(anotherValue),
		AvailableGas: 200000,
		GasPrice:     (*hexutil.Big)(big.NewInt(1000000000)),
	})
	// call function to retrieve value
	resultRetrieve := call[lib.EvmResult](t, instance, "EvmApply", lib.EvmParams{
		HandleParams: lib.HandleParams{Handle: handle},
		From:         user,
		To:           result.ContractAddress,
		Input:        Storage.Retrieve(),
		AvailableGas: 200000,
		GasPrice:     (*hexutil.Big)(big.NewInt(1000000000)),
	})
	if resultRetrieve.EvmError != "" {
		t.Fatalf("vm error: %v", resultRetrieve.EvmError)
	}
	retrievedValue := common.BytesToHash(resultRetrieve.ReturnData).Big()
	if anotherValue.Cmp(retrievedValue) != 0 {
		t.Fatalf("retrieved bad value: expected %v, actual %v", anotherValue, retrievedValue)
	}
	// verify that EOA nonce was not updated
	nonce := call[hexutil.Uint64](t, instance, "StateGetNonce", lib.AccountParams{
		HandleParams: lib.HandleParams{Handle: handle},
		Address:      user,
	})
	if uint64(nonce) != 1 {
		t.Fatalf("nonce was modified: expected 0, actual %v", nonce)
	}
	// cleanup
	call[any](t, instance, "CloseDatabase", lib.DatabaseParams{DatabaseHandle: dbHandle})
}
