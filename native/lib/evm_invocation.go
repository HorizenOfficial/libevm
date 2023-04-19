package lib

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
)

type Invocation struct {
	Caller   common.Address `json:"caller"`
	Callee   common.Address `json:"callee"`
	Value    *hexutil.Big   `json:"value"`
	Input    []byte         `json:"input"`
	Gas      hexutil.Uint64 `json:"gas"`
	ReadOnly bool           `json:"readOnly"`
}

type InvocationResult struct {
	ReturnData     []byte         `json:"returnData"`
	LeftOverGas    hexutil.Uint64 `json:"leftOverGas"`
	ExecutionError error          `json:"executionError"`
}

type InvocationCallback struct{ Callback }

func (c *InvocationCallback) execute(caller, callee common.Address, value *big.Int, input []byte, gas uint64, readOnly bool) (ret []byte, leftOverGas uint64, err error) {
	if c == nil {
		// fallback to noop
		return nil, gas, nil
	}
	invocation := &Invocation{
		Caller:   caller,
		Callee:   callee,
		Value:    (*hexutil.Big)(value),
		Input:    input,
		Gas:      hexutil.Uint64(gas),
		ReadOnly: readOnly,
	}
	result := new(InvocationResult)
	err = c.Invoke(invocation, result)
	if err != nil {
		log.Error("block hash getter callback failed: %v", err)
		return nil, gas, err
	}
	return result.ReturnData, uint64(result.LeftOverGas), result.ExecutionError
}
