package lib

import (
	"fmt"
	"github.com/HorizenOfficial/go-ethereum/common"
	"github.com/HorizenOfficial/go-ethereum/common/hexutil"
	"libevm/interop"
	"libevm/test"
	"math/big"
	"reflect"
	"strings"
	"testing"
)

func TestEvmOpCodes(t *testing.T) {
	var (
		instance, _, stateHandle = SetupTest()
		user                     = common.HexToAddress("0x42")
		_, statedb               = instance.statedbs.Get(stateHandle)
	)

	// deploy "OpCodes" contract
	_, resultDeploy := instance.EvmApply(EvmParams{
		HandleParams: HandleParams{Handle: stateHandle},
		Invocation: Invocation{
			Caller: user,
			Callee: nil,
			Input:  test.OpCodes.Deploy(),
			Gas:    200000,
		},
		Context: EvmContext{Rules: &ForkRules{IsShanghai: true}},
	})
	if resultDeploy.ExecutionError != "" {
		t.Fatalf("vm error: %v", resultDeploy.ExecutionError)
	}
	deployedCode := statedb.GetCode(*resultDeploy.ContractAddress)
	if common.Bytes2Hex(test.OpCodes.RuntimeCode()) != common.Bytes2Hex(deployedCode) {
		t.Fatalf("deployed code does not match %s", common.Bytes2Hex(deployedCode))
	}

	var (
		gasPrice    = big.NewInt(586732)
		chainID     = uint64(1337)
		coinbase    = common.HexToAddress("0x09a1e4d0c6f6055287a6e1553a1d9cfe05767591")
		gasLimit    = uint64(32123457)
		blockNumber = big.NewInt(51231287)
		time        = big.NewInt(1669144595)
		baseFee     = big.NewInt(123872)
		random      = common.HexToHash("0x0a5d85d0f0e021c04643e05e38f8f28029275683ee743910670154d78322b6eb")
		blockHash   = common.HexToHash("0xc01a0d15649a201418433e1760af47a0c3381bc7aec566f1e6258d77ffd2e2c9")
	)

	// redefine this interface here, because it is not exported from GETH
	type bytesBacked interface {
		Bytes() []byte
	}

	// setup callback proxy for the BLOCKHASH opcode
	const blockHashCallbackHandle = 5132
	SetCallbackProxy(func(handle int, args string) string {
		switch handle {
		case blockHashCallbackHandle:
			actual := new(hexutil.Big)
			err := interop.Deserialize(args, actual)
			if err != nil {
				panic(fmt.Sprintf("invalid callback arguments: %v", args))
			}
			// the getBlockHash() function of OpCodes.sol will always call blockhash of blockNumber - 1
			// verify that the argument arrived here as expected
			if expected := new(big.Int).Sub(blockNumber, common.Big1); actual.ToInt().Cmp(expected) != 0 {
				panic(fmt.Sprintf("BLOCKHASH opcode called with unexpected block number: want %v got %v", expected, actual))
			}
			result, _ := interop.Serialize(blockHash)
			return result
		default:
			panic(fmt.Sprintf("callback proxy called with unknown handle: %v args: %s", handle, args))
		}
	})

	checks := []struct {
		name     string
		expected bytesBacked
	}{
		{"GASPRICE", gasPrice},
		{"CHAINID", new(big.Int).SetUint64(chainID)},
		{"COINBASE", coinbase},
		{"GASLIMIT", new(big.Int).SetUint64(gasLimit)},
		{"BLOCKNUMBER", blockNumber},
		{"TIME", time},
		{"BASEFEE", baseFee},
		{"RANDOM", random},
		{"BLOCKHASH", blockHash},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			// call function and verify result value
			err, result := instance.EvmApply(EvmParams{
				HandleParams: HandleParams{Handle: stateHandle},
				Invocation: Invocation{
					Caller: user,
					Callee: resultDeploy.ContractAddress,
					Input:  test.OpCodes.Call(check.name),
					Gas:    200000,
				},
				Context: EvmContext{
					ChainID:           hexutil.Uint64(chainID),
					Coinbase:          coinbase,
					GasLimit:          hexutil.Uint64(gasLimit),
					GasPrice:          (*hexutil.Big)(gasPrice),
					BlockNumber:       (*hexutil.Big)(blockNumber),
					Time:              (*hexutil.Big)(time),
					BaseFee:           (*hexutil.Big)(baseFee),
					Random:            random,
					BlockHashCallback: &BlockHashCallback{Callback(blockHashCallbackHandle)},
					Rules:             &ForkRules{IsShanghai: true},
				},
			})
			if err != nil {
				t.Fatalf("error: %v", result.ExecutionError)
			}
			if result.ExecutionError != "" {
				t.Fatalf("vm error: %v", result.ExecutionError)
			}
			if expected := common.LeftPadBytes(check.expected.Bytes(), 32); !reflect.DeepEqual(expected, result.ReturnData) {
				t.Fatalf("test failed for %v:\n%v expected\n%v actual", check.name, expected, result.ReturnData)
			}
		})
	}
}

func TestEvmExternalContracts(t *testing.T) {
	var (
		instance, _, stateHandle    = SetupTest()
		handle                      = HandleParams{Handle: stateHandle}
		user                        = common.HexToAddress("0xbafe3b6f2a19658df3cb5efca158c93272ff5c0b")
		forgerStakesContractAddress = common.HexToAddress("0x0000000000000000000022222222222222222222")
		mockedForgerStakesData      = common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000bc22971e6a19a3ddf28dafcbe3eaf261cfac0f3dc07f9cef79dfc94175d1eb8cc000000000000000000000000000000000000000000000000016345785d8a000000000000000000000000000039ef4eea2b229c11dba18392b7f12c83c89053b9ffc4065704d28337afdb9e216e264be8e1121f29b33a84860b78008cf8638ebe0bfce2073aa51bc5854f8f297e31b89358a29e27f95380b91268e68dc58d1d0a8000000000000000000000000000000000000000000000000000000000000000fe9436b7f4645cc5562e7f37996a11a63da703043df985ec23f9c5a642a288ff000000000000000000000000000000000000000000000000016345785d8a00000000000000000000000000002a789f245142753075c4c5a3c603b52d8f01d361b85bb1e0b3de12c8bb3b7f7cb04ee5f7cdec4de4bf0879536824cc84adf6d7d29e9f78162e4509a0df40e3d8525801c363a8d327ce0c9eb4da8a3615489aee1680000000000000000000000000000000000000000000000000000000000000000e19a05f03af33541c3219a734b0868f39b36493b3e8ba63c86c5659725d582c000000000000000000000000000000000000000000000000016345785d8a0000000000000000000000000000551c0294ef40e7d4da0d0e62f77141cdb455dcc47b48711bc08c786c486520bb9eb69ba43ed04fc2cc617c24e85d757ffcaf9a971644ba05f4fcea365fabba5b1ee38f0a1db512f7178c3733e8680589cb4d7f350000000000000000000000000000000000000000000000000000000000000000aa27870083d34abec0abf7a0a5e39d2cb353ffb96c486af53868783147211f18000000000000000000000000000000000000000000000000016345785d8a000000000000000000000000000072b6b2ed1c4d951a5247206b771a2615ec25a255087a518986785e8746d114d6cbdea0fc9f048362b1b318b4d36f7d1d51596ac789c3617a58b703631dcff9ce96697e0b4e1c7ce9207db22edf110e5b29347c1580000000000000000000000000000000000000000000000000000000000000001ddda8086cbc1d0752e95651ba810b7a56441bf42a770010475f8d02774f7411000000000000000000000000000000000000000000000000016345785d8a0000000000000000000000000000e36c4e40d357e4a6e4df7a05f5944264df867265c0905c48964a940880e8de76c84dddc8535d0878a693097b3e86f6509e649606b5c5916187e905d83ba75ebdc39d2d526d6402a4801677f325dd55e1df5d810d000000000000000000000000000000000000000000000000000000000000000086c07dadb86c083eb56e6e9ac3891afab10806b8221b50a5763ad56748b51113000000000000000000000000000000000000000000000000016345785d8a00000000000000000000000000002e0126cbcae91e490522f65925773d182339799674cac9966adc776d28720f7bfa744e7c725b7388734bb6801b98f3470aefdf3f0ce9e0bd37fb0abcc3dc392baeeb8ac3acd47ac73ae79429cd9f4807faec06330000000000000000000000000000000000000000000000000000000000000000fcbd1f41cf3f859f0f3c3de58d4227032400e9edc46feb81ad0d9790283be923000000000000000000000000000000000000000000000000016345785d8a00000000000000000000000000008bf4857b740943b005fbbee80719421ab5b3dc510ef5b3566378c969af6545b06706e380cba6ba6dc55d7c1a5ae0c914475100c7c7b855d86e93a45d0108026614c4a8e2acb201f7c4e0f26297e3fbde0cf57e268000000000000000000000000000000000000000000000000000000000000000703bf3e9519b03c8e97f29d5e4cce7e136f894a6fd5eb35020f193ffb77b838c000000000000000000000000000000000000000000000000016345785d8a000000000000000000000000000008a698b7535d6eeb6341e5c65e220c49f177f73e6cd00ba09b5d2b771c1f90399e183bf80e54886a390904be2e018d0d80b46b018d71700c48baf671f9ebd48895f555f32b7b0c20ee2a3dcb35290c9a412a2e03800000000000000000000000000000000000000000000000000000000000000006f64a8254e75c9dd9dba6499112c0fe99b7ad5f1d978ace402c3459d1d09891000000000000000000000000000000000000000000000000016345785d8a000000000000000000000000000030d617085728a81563dfd655c35020ab8b78ecec76f4fb6b460dbc0fdf3e17fb5b9888ab9ce2f2f009e2fb32ef95eff196b4326eacde2c72c65e185d9e04e6012ce0bc6165a17a4f95215d701982c8b8cb5d970800000000000000000000000000000000000000000000000000000000000000008d546419e013bafbaead980365902a624f724cd1611086ab55946ec72cbfe889000000000000000000000000000000000000000000000000016345785d8a0000000000000000000000000000e089abfcb53895adcbc8825aba391ebd1ddd83ed12e73f518bd4314b5251d881394711433f348ebea66b817aade3c93ffb3650dc51c95d2cd6428da7189c3e7451e51906bde78fee932f3a61644c22ca8856f0320000000000000000000000000000000000000000000000000000000000000000fda9653e35779b6b8f5186548d9bd3e7f66f6bc662d530569104cc1ad6afe437000000000000000000000000000000000000000000000000016345785d8a000000000000000000000000000075413637bc6e1818b36c9acf56fac28765bc3ec8fa9471609398bc7520e8151a96409151433d426a094fc70e6f32b89ed854d12acd7a2250e51d6ef985275847e7c715e67b4967a44910d19950c19bf43578a3210000000000000000000000000000000000000000000000000000000000000000")
	)

	_, deployResult := instance.EvmApply(EvmParams{
		HandleParams: HandleParams{Handle: stateHandle},
		Invocation: Invocation{
			Caller: user,
			Input:  test.NativeInterop.Deploy(),
			Gas:    500000,
		},
		Context: EvmContext{Rules: &ForkRules{IsShanghai: true}},
	})
	if deployResult.ExecutionError != "" {
		t.Fatalf("vm error: %v", deployResult.ExecutionError)
	}

	const invocationCallbackHandle = 128967
	expectedInvocation := &ExternalInvocation{
		Invocation: Invocation{
			Caller:   *deployResult.ContractAddress,
			Callee:   &forgerStakesContractAddress,
			Value:    (*hexutil.Big)(common.Big0),
			Input:    test.ForgerStakes.GetAllForgersStakes(),
			Gas:      10000,
			ReadOnly: true,
		},
		Depth: 1,
	}
	SetCallbackProxy(func(handle int, args string) string {
		switch handle {
		case invocationCallbackHandle:
			actual := new(ExternalInvocation)
			err := interop.Deserialize(args, actual)
			if err != nil {
				panic(fmt.Sprintf("invalid callback arguments: %v", args))
			}
			if expectedArgs, _ := interop.Serialize(expectedInvocation); expectedArgs != args {
				panic(fmt.Sprintf("unexpected invocation arguments:\nwant %v\n got %v", expectedArgs, args))
			}
			res := InvocationResult{
				ReturnData:     mockedForgerStakesData,
				LeftOverGas:    0,
				ExecutionError: "",
			}
			result, _ := interop.Serialize(res)
			return result
		default:
			panic(fmt.Sprintf("callback proxy called with unknown handle: %v args: %s", handle, args))
		}
	})

	_, callResult := instance.EvmApply(EvmParams{
		HandleParams: handle,
		Invocation: Invocation{
			Caller: user,
			Callee: deployResult.ContractAddress,
			Input:  test.NativeInterop.GetForgerStakes(),
			Gas:    200000,
		},
		Context: EvmContext{
			ExternalContracts: []common.Address{forgerStakesContractAddress},
			ExternalCallback:  &InvocationCallback{Callback(invocationCallbackHandle)},
			InitialDepth:      20,
			Rules:             &ForkRules{IsShanghai: true},
		},
	})
	if callResult.ExecutionError != "" {
		t.Fatalf("vm error: %v", callResult.ExecutionError)
	}

	_, delegateCallResult := instance.EvmApply(EvmParams{
		HandleParams: handle,
		Invocation: Invocation{
			Caller: user,
			Callee: deployResult.ContractAddress,
			Input:  test.NativeInterop.GetForgerStakesDelegateCall(),
			Gas:    200000,
		},
		Context: EvmContext{
			ExternalContracts: []common.Address{forgerStakesContractAddress},
			ExternalCallback:  &InvocationCallback{Callback(invocationCallbackHandle)},
			Rules:             &ForkRules{IsShanghai: true},
		},
	})
	if !strings.Contains(delegateCallResult.ExecutionError, "unsupported call method") {
		t.Fatal("expected vm error on DelegateCall")
	}
}

func TestEvmErrors(t *testing.T) {
	var (
		instance, _, stateHandle = SetupTest()
		_, statedb               = instance.statedbs.Get(stateHandle)
		user                     = common.HexToAddress("0xbafe3b6f2a19658df3cb5efca158c93272ff5c0b")
		addr                     = common.HexToAddress("0x1234")
	)

	// deploy test contract
	_, resultDeploy := instance.EvmApply(EvmParams{
		HandleParams: HandleParams{Handle: stateHandle},
		Invocation: Invocation{
			Caller: user,
			Callee: nil,
			Input:  test.Storage.Deploy(common.Big0),
			Gas:    200000,
		},
		Context: EvmContext{Rules: &ForkRules{IsShanghai: true}},
	})
	if resultDeploy.ExecutionError != "" {
		t.Fatalf("vm error: %v", resultDeploy.ExecutionError)
	}

	// add some test balance
	statedb.AddBalance(user, big.NewInt(1000))

	checks := []struct {
		name         string
		err          string
		shouldRevert bool
		params       EvmParams
	}{
		{
			name: "EOA transfer with insufficient balance",
			err:  "insufficient balance for transfer",
			params: EvmParams{
				HandleParams: HandleParams{Handle: stateHandle},
				Invocation: Invocation{
					Caller: user,
					Callee: &addr,
					Value:  (*hexutil.Big)(big.NewInt(1001)),
					Gas:    100,
				},
				Context: EvmContext{Rules: &ForkRules{IsShanghai: true}},
			},
		},
		{
			name: "contract deployment with insufficient gas for execution",
			err:  "out of gas",
			params: EvmParams{
				HandleParams: HandleParams{Handle: stateHandle},
				Invocation: Invocation{
					Caller: user,
					Input:  test.Storage.Deploy(common.Big0),
					Gas:    123,
				},
				Context: EvmContext{Rules: &ForkRules{IsShanghai: true}},
			},
		},
		{
			name: "contract deployment with insufficient gas for code storage",
			err:  "contract creation code storage out of gas",
			params: EvmParams{
				HandleParams: HandleParams{Handle: stateHandle},
				Invocation: Invocation{
					Caller: user,
					Input:  test.Storage.Deploy(common.Big0),
					Gas:    50000,
				},
				Context: EvmContext{Rules: &ForkRules{IsShanghai: true}},
			},
		},
		{
			// illegal because the constructor of this contract does not accept funds
			name:         "contract deployment with illegal value transfer",
			err:          "execution reverted",
			shouldRevert: true,
			params: EvmParams{
				HandleParams: HandleParams{Handle: stateHandle},
				Invocation: Invocation{
					Caller: user,
					Value:  (*hexutil.Big)(big.NewInt(100)),
					Input:  test.Storage.Deploy(common.Big0),
					Gas:    200000,
				},
				Context: EvmContext{Rules: &ForkRules{IsShanghai: true}},
			},
		},
		{
			name:         "contract call to unknown function",
			err:          "execution reverted",
			shouldRevert: true,
			params: EvmParams{
				HandleParams: HandleParams{Handle: stateHandle},
				Invocation: Invocation{
					Caller: user,
					Callee: resultDeploy.ContractAddress,
					Input:  common.FromHex("01020304"),
					Gas:    200000,
				},
				Context: EvmContext{Rules: &ForkRules{IsShanghai: true}},
			},
		},
		{
			name: "contract call with insufficient gas",
			err:  "out of gas",
			params: EvmParams{
				HandleParams: HandleParams{Handle: stateHandle},
				Invocation: Invocation{
					Caller: user,
					Callee: resultDeploy.ContractAddress,
					Input:  test.Storage.Store(common.Big3),
					Gas:    2000,
				},
				Context: EvmContext{Rules: &ForkRules{IsShanghai: true}},
			},
		},
		{
			name: "stateful contract call with read-only enabled",
			err:  "write protection",
			params: EvmParams{
				HandleParams: HandleParams{Handle: stateHandle},
				Invocation: Invocation{
					Caller:   user,
					Callee:   resultDeploy.ContractAddress,
					Input:    test.Storage.Store(common.Big3),
					Gas:      200000,
					ReadOnly: true,
				},
				Context: EvmContext{Rules: &ForkRules{IsShanghai: true}},
			},
		},
	}

	for i, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			// make sure the nonce is increased between calls, otherwise there might be contract address collisions
			statedb.SetNonce(user, uint64(i+1))
			err, result := instance.EvmApply(check.params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.ExecutionError != check.err {
				t.Fatalf("unexpected ExecutionError: expected \"%v\" actual \"%v\"", check.err, result.ExecutionError)
			}
		})
	}
}

func TestShanghaiFork(t *testing.T) {
	var (
		instance, _, stateHandle = SetupTest()
		user                     = common.HexToAddress("0x42")
	)

	// Test that deploying a contract compiled for Shanghai fails if fork is not active

	// deploy "OpCodes" contract
	_, resultDeploy := instance.EvmApply(EvmParams{
		HandleParams: HandleParams{Handle: stateHandle},
		Invocation: Invocation{
			Caller: user,
			Callee: nil,
			Input:  test.OpCodes.Deploy(),
			Gas:    200000,
		},
		Context: EvmContext{Rules: &ForkRules{IsShanghai: false}},
	})
	if resultDeploy.ExecutionError != "invalid opcode: PUSH0" {
		t.Fatalf("vm error: %v", resultDeploy.ExecutionError)
	}

	// Same with Shanghai active, the deployment should succeed

	// deploy "OpCodes" contract
	_, resultDeploy = instance.EvmApply(EvmParams{
		HandleParams: HandleParams{Handle: stateHandle},
		Invocation: Invocation{
			Caller: user,
			Callee: nil,
			Input:  test.OpCodes.Deploy(),
			Gas:    200000,
		},
		Context: EvmContext{Rules: &ForkRules{IsShanghai: true}},
	})
	if resultDeploy.ExecutionError != "" {
		t.Fatalf("vm error: %v", resultDeploy.ExecutionError)
	}

}
