package test

import (
	"encoding/json"
	"github.com/HorizenOfficial/go-ethereum/common"
	"github.com/HorizenOfficial/go-ethereum/common/hexutil"
	"libevm/interop"
	"libevm/lib"
	"math/big"
	"testing"
)

func call[T any](t *testing.T, instance *lib.Service, method string, args any) T {
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

	dbHandle := call[int](t, instance, "DatabaseOpenLevelDB", lib.LevelDBParams{Path: t.TempDir()})
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
	result := call[lib.InvocationResult](t, instance, "EvmApply", lib.EvmParams{
		HandleParams: lib.HandleParams{Handle: handle},
		Invocation: lib.Invocation{
			Caller: user,
			Callee: nil,
			Input:  Storage.Deploy(initialValue),
			Gas:    200000,
		},
	})
	if result.ExecutionError != "" {
		t.Fatalf("vm error: %v", result.ExecutionError)
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
		Invocation: lib.Invocation{
			Caller: user,
			Callee: result.ContractAddress,
			Input:  Storage.Store(anotherValue),
			Gas:    200000,
		},
	})
	// call function to retrieve value
	resultRetrieve := call[lib.InvocationResult](t, instance, "EvmApply", lib.EvmParams{
		HandleParams: lib.HandleParams{Handle: handle},
		Invocation: lib.Invocation{
			Caller: user,
			Callee: result.ContractAddress,
			Input:  Storage.Retrieve(),
			Gas:    200000,
		},
	})
	if resultRetrieve.ExecutionError != "" {
		t.Fatalf("vm error: %v", resultRetrieve.ExecutionError)
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
	call[any](t, instance, "DatabaseClose", lib.DatabaseParams{DatabaseHandle: dbHandle})
}
