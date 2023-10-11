package lib

import (
	"fmt"
	"github.com/HorizenOfficial/go-ethereum/common"
	"github.com/HorizenOfficial/go-ethereum/common/hexutil"
	"github.com/HorizenOfficial/go-ethereum/core/vm"
	"github.com/HorizenOfficial/go-ethereum/log"
	"math/big"
)

type InvocationCallback struct{ Callback }

type ExternalInvocation struct {
	Invocation
	Depth int `json:"depth"`
}

func (c *InvocationCallback) execute(caller, callee common.Address, value *big.Int, input []byte, gas uint64, readOnly bool, depth int) (ret []byte, leftOverGas uint64, err error) {
	if c == nil {
		// fallback to noop
		return nil, gas, nil
	}
	invocation := &ExternalInvocation{
		Invocation: Invocation{
			Caller:   caller,
			Callee:   &callee,
			Value:    (*hexutil.Big)(value),
			Input:    input,
			Gas:      hexutil.Uint64(gas),
			ReadOnly: readOnly,
		},
		Depth: depth,
	}
	result := new(InvocationResult)
	err = c.Invoke(invocation, result)
	if err != nil {
		log.Error("invocation callback failed", "err", err)
		return nil, gas, err
	}
	var invocationErr error
	if result.Reverted {
		invocationErr = vm.ErrExecutionReverted
	} else if result.ExecutionError != "" {
		invocationErr = fmt.Errorf("external contract invocation failed: %s", result.ExecutionError)
	}
	return result.ReturnData, uint64(result.LeftOverGas), invocationErr
}
