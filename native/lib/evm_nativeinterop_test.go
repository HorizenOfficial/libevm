package lib

import (
	"github.com/ethereum/go-ethereum/common"
	"libevm/test"
	"testing"
)

func TestNativeInterop(t *testing.T) {
	var (
		instance, _, stateHandle = SetupTest()
		handle                   = HandleParams{Handle: stateHandle}
		user                     = common.HexToAddress("0xbafe3b6f2a19658df3cb5efca158c93272ff5c0b")
	)

	_, deployResult := instance.EvmApply(EvmParams{
		HandleParams: HandleParams{Handle: stateHandle},
		From:         user,
		Input:        test.NativeInterop.Deploy(),
		AvailableGas: 200000,
	})
	if deployResult.EvmError != "" {
		t.Fatalf("vm error: %v", deployResult.EvmError)
	}

	_, callResult := instance.EvmApply(EvmParams{
		HandleParams: handle,
		From:         user,
		To:           deployResult.ContractAddress,
		Input:        test.NativeInterop.GetForgerStakes(),
		AvailableGas: 200000,
	})
	if callResult.EvmError != "" {
		t.Fatalf("vm error: %v", callResult.EvmError)
	}

}
