package lib

import (
	"encoding/json"
	"libevm/test"
	"math"
	"math/big"
	"testing"

	"github.com/HorizenOfficial/go-ethereum/common"
	"github.com/HorizenOfficial/go-ethereum/common/hexutil"
	"github.com/HorizenOfficial/go-ethereum/eth/tracers/logger"
)

func getTracerResult(t *testing.T, params TracerCreateParams) *TracerResult {
	var (
		instance, _, stateHandle = SetupTest()
		initialValue             = common.Big0
		sender                   = common.HexToAddress("0xbafe3b6f2a19658df3cb5efca158c93272ff5c0b")
	)
	err, tracerHandle := instance.TracerCreate(params)
	if err != nil {
		t.Fatal(err)
	}
	err, result := instance.EvmApply(EvmParams{
		HandleParams: HandleParams{
			Handle: stateHandle,
		},
		Invocation: Invocation{
			Caller: sender,
			Callee: nil,
			Input:  test.Storage.Deploy(initialValue),
			Gas:    (hexutil.Uint64)(math.MaxInt64),
		},
		Context: EvmContext{
			BaseFee: (*hexutil.Big)(new(big.Int)),
			Tracer:  &tracerHandle,
			Rules:   &ForkRules{IsShanghai: true},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if result.Reverted || result.ExecutionError != "" {
		t.Fatal(result)
	}
	// retrieve the result and check that its structure is of type logger.ExecutionResult
	err, tracerResult := instance.TracerResult(TracerParams{TracerHandle: tracerHandle})
	if err != nil {
		t.Fatal(err)
	}
	instance.TracerRemove(TracerParams{TracerHandle: tracerHandle})
	return tracerResult
}

func TestTracer_StructLogger(t *testing.T) {
	tracerResult := getTracerResult(t, TracerCreateParams{
		EnableMemory:     true,
		DisableStack:     false,
		DisableStorage:   false,
		EnableReturnData: true,
	})
	// retrieve the result and check that its structure is of type logger.ExecutionResult
	var trace *logger.ExecutionResult
	if err := json.Unmarshal(tracerResult.Result, &trace); err != nil {
		t.Fatal(err)
	}
	// do a coarse correctness check that does not immediately break on different versions of the solidity compiler
	if minimum, actual := 130, len(trace.StructLogs); minimum > actual {
		t.Fatalf("unexpected number of trace logs: expected at least %d, actual %d", minimum, actual)
	}
	// cherry-pick the one SSTORE instruction that should be in there
	sstoreInstructions := 0
	for _, trace := range trace.StructLogs {
		if trace.Op != "SSTORE" {
			continue
		}
		sstoreInstructions += 1
		if expected, actual := "SSTORE", trace.Op; expected != actual {
			t.Fatalf("unexpected op code: expected %s, actual %s", expected, actual)
		}
		if expected, actual := 1, len(*trace.Storage); expected != actual {
			t.Fatalf("unexpected number of accessed storage keys: expected %d, actual %d", expected, actual)
		}
	}
	if sstoreInstructions != 1 {
		t.Fatalf("unexpected number of SSTORE instructions: expected %d, actual %d", 1, sstoreInstructions)
	}
}

// CallTracer logger response structure used in tests
type CallTracer struct {
	Type    string
	From    string
	To      string
	Value   string
	Gas     string
	GasUsed string
	Input   string
	Output  string
	Error   string
	Calls   []CallTracer
}

func TestTracer_CallTracer(t *testing.T) {
	tracerResult := getTracerResult(t, TracerCreateParams{
		Tracer: "callTracer",
	})
	// retrieve the result and check that its structure is of type CallFrame
	var trace *CallTracer
	if err := json.Unmarshal(tracerResult.Result, &trace); err != nil {
		t.Fatal(err)
	}
	if trace.Type != "CREATE" &&
		trace.From != "0xbafe3b6f2a19658df3cb5efca158c93272ff5c0b" &&
		trace.To != "0x6f8c38b30df9967a414543a1338d4497f2570775" {
		t.Fatalf("unexpected trace: %v", trace)
	}
}

func TestTracer_CallTracerWithTracerConfig(t *testing.T) {
	tracerResult := getTracerResult(t, TracerCreateParams{
		Tracer:       "callTracer",
		TracerConfig: json.RawMessage(`{"onlyTopCall": true, "withLog": false}`),
	})
	// retrieve the result and check that its structure is of type CallFrame
	var trace *CallTracer
	if err := json.Unmarshal(tracerResult.Result, &trace); err != nil {
		t.Fatal(err)
	}
	if trace.Type != "CREATE" &&
		trace.From != "0xbafe3b6f2a19658df3cb5efca158c93272ff5c0b" &&
		trace.To != "0x6f8c38b30df9967a414543a1338d4497f2570775" {
		t.Fatalf("unexpected trace: %v", trace)
	}
}

func TestTracer_FourByteTracer(t *testing.T) {
	tracerResult := getTracerResult(t, TracerCreateParams{
		Tracer:       "4byteTracer",
		TracerConfig: json.RawMessage(`{"onlyTopCall": true, "withLog": false}`),
	})
	// retrieve the result and check that its structure is a string-int map
	var trace map[string]int
	if err := json.Unmarshal(tracerResult.Result, &trace); err != nil {
		t.Fatal(err)
	}
}
