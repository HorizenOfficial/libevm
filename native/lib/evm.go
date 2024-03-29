package lib

import (
	"math"
	"math/big"
	"time"

	"github.com/HorizenOfficial/go-ethereum/common"
	"github.com/HorizenOfficial/go-ethereum/common/hexutil"
	"github.com/HorizenOfficial/go-ethereum/core"
	"github.com/HorizenOfficial/go-ethereum/core/state"
	"github.com/HorizenOfficial/go-ethereum/core/vm"
	"github.com/HorizenOfficial/go-ethereum/eth/tracers"
	"github.com/HorizenOfficial/go-ethereum/params"
)

type ForkRules struct {
	IsShanghai bool `json:"isShanghai"`
}

type Invocation struct {
	Caller   common.Address  `json:"caller"`
	Callee   *common.Address `json:"callee"`
	Value    *hexutil.Big    `json:"value"`
	Input    []byte          `json:"input"`
	Gas      hexutil.Uint64  `json:"gas"`
	ReadOnly bool            `json:"readOnly"`
}

type InvocationResult struct {
	ReturnData      []byte          `json:"returnData"`
	LeftOverGas     hexutil.Uint64  `json:"leftOverGas"`
	ExecutionError  string          `json:"executionError"`
	Reverted        bool            `json:"reverted"`
	ContractAddress *common.Address `json:"contractAddress"`
}

type EvmContext struct {
	ChainID           hexutil.Uint64      `json:"chainID"`
	Coinbase          common.Address      `json:"coinbase"`
	GasLimit          hexutil.Uint64      `json:"gasLimit"`
	GasPrice          *hexutil.Big        `json:"gasPrice"`
	BlockNumber       *hexutil.Big        `json:"blockNumber"`
	Time              *hexutil.Big        `json:"time"`
	BaseFee           *hexutil.Big        `json:"baseFee"`
	Random            common.Hash         `json:"random"`
	BlockHashCallback *BlockHashCallback  `json:"blockHashCallback"`
	Tracer            *int                `json:"tracer"`
	ExternalContracts []common.Address    `json:"externalContracts"`
	ExternalCallback  *InvocationCallback `json:"externalCallback"`
	InitialDepth      int                 `json:"initialDepth"`
	Rules             *ForkRules          `json:"rules"`
}

// setDefaults for parameters that were omitted
func (p *Invocation) setDefaults() {
	if p.Value == nil {
		p.Value = (*hexutil.Big)(common.Big0)
	}
}

// setDefaults for parameters that were omitted
func (c *EvmContext) setDefaults() {
	if c.GasLimit == 0 {
		c.GasLimit = (hexutil.Uint64)(math.MaxInt64)
	}
	if c.GasPrice == nil {
		c.GasPrice = (*hexutil.Big)(common.Big0)
	}
	if c.BlockNumber == nil {
		c.BlockNumber = (*hexutil.Big)(common.Big0)
	}
	if c.Time == nil {
		c.Time = (*hexutil.Big)(big.NewInt(time.Now().Unix()))
	}
	if c.BaseFee == nil {
		c.BaseFee = (*hexutil.Big)(big.NewInt(params.InitialBaseFee))
	}
	if c.Rules == nil {
		c.Rules = &ForkRules{IsShanghai: false}
	}
}

func (c *EvmContext) getBlockContext() vm.BlockContext {
	return vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     c.BlockHashCallback.getBlockHash,
		Coinbase:    c.Coinbase,
		GasLimit:    uint64(c.GasLimit),
		BlockNumber: c.BlockNumber.ToInt(),
		Time:        c.Time.ToInt().Uint64(),
		Difficulty:  common.Big0,
		BaseFee:     c.BaseFee.ToInt(),
		Random:      &c.Random,
	}
}

func (c *EvmContext) getChainConfig() *params.ChainConfig {
	var shanghaiTime *uint64 = nil //When nil means is disabled (see "github.com/HorizenOfficial/go-ethereum/params/config.go")
	if c.Rules.IsShanghai {
		shanghaiTime = new(uint64) //When 0 means is enabled (see "github.com/HorizenOfficial/go-ethereum/params/config.go")
	}
	return &params.ChainConfig{
		ChainID:             new(big.Int).SetUint64(uint64(c.ChainID)),
		HomesteadBlock:      common.Big0,
		DAOForkBlock:        nil,
		DAOForkSupport:      false,
		EIP150Block:         common.Big0,
		EIP155Block:         common.Big0,
		EIP158Block:         common.Big0,
		ByzantiumBlock:      common.Big0,
		ConstantinopleBlock: common.Big0,
		PetersburgBlock:     common.Big0,
		IstanbulBlock:       common.Big0,
		MuirGlacierBlock:    common.Big0,
		BerlinBlock:         common.Big0,
		LondonBlock:         common.Big0,
		ShanghaiTime:        shanghaiTime,
	}
}

// getTracer retrieves an instance of a tracer if a handle is given
func (c *EvmContext) getTracer(s *Service) (error, tracers.Tracer) {
	if c.Tracer == nil {
		return nil, nil
	}
	err, tracerPtr := s.tracers.Get(*c.Tracer)
	if err != nil {
		return err, nil
	}
	return nil, *tracerPtr
}

type EvmParams struct {
	HandleParams
	Invocation Invocation `json:"invocation"`
	Context    EvmContext `json:"context"`
}

func (s *Service) getEvm(context EvmContext, stateDB *state.StateDB, origin common.Address) (error, *vm.EVM) {
	// get tracer if a handle is given
	err, tracer := context.getTracer(s)
	if err != nil {
		return err, nil
	}

	var (
		txContext = vm.TxContext{
			Origin:   origin,
			GasPrice: context.GasPrice.ToInt(),
		}
		blockContext = context.getBlockContext()
		chainConfig  = context.getChainConfig()
		evmConfig    = vm.Config{
			Tracer:                  tracer,
			NoBaseFee:               false,
			EnablePreimageRecording: false,
			ExtraEips:               nil,
			InitialDepth:            context.InitialDepth,
			ExternalContracts:       context.ExternalContracts,
			ExternalCallback:        context.ExternalCallback.execute,
		}
	)
	return nil, vm.NewEVM(blockContext, txContext, stateDB, chainConfig, evmConfig)
}

func (s *Service) EvmApply(params EvmParams) (error, *InvocationResult) {
	// apply defaults to missing parameters
	params.Invocation.setDefaults()
	params.Context.setDefaults()

	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err, nil
	}

	err, evm := s.getEvm(params.Context, statedb, params.Invocation.Caller)
	if err != nil {
		return err, nil
	}

	var (
		invocation       = params.Invocation
		sender           = vm.AccountRef(invocation.Caller)
		gas              = uint64(invocation.Gas)
		contractCreation = invocation.Callee == nil
	)

	var (
		returnData      []byte
		vmerr           error
		contractAddress *common.Address
	)
	if contractCreation {
		// The following nonce modification is a workaround for the following problem:
		//
		// Creating a smart contract should increment the callers' nonce, this is true for EOAs as well as contracts
		// creating other contracts. Thus, the nonce increment is done in evm.Create and must be there.
		// In contrast to that behavior, for the top level call the nonce was already increased by the SDK at this
		// point. So if we don't do anything here the nonce of an EOA will be increased twice when a smart contract is
		// deployed.
		//
		// As the contract address is calculated from the nonce we can't just decrement the nonce afterwards (to undo
		// the unwanted change), we have to do that before running the EVM. This also introduces two edge cases:
		//
		// - The check nonce > 0 was necessary in an earlier version where the nonce was NOT increased when the call was
		//   performed in the context of eth_call via RPC. This is fixed now, but we should still keep this as a
		//   precaution (this would cause unsigned integer underflow to maxUint64) and because it is useful for tests.
		//
		// - The EVM.create call can fail before it even reaches the point of incrementing the nonce. We have to make
		//   sure to NOT decrement the nonce in that case. Hence, setting the nonce to the value before the EVM call in
		//   case it was modified.
		// This applies only to top level calls. It may happen that a contract creation is requested by a native smart contract.
		// In this case the nonce of the native smart contract is not incremented before this call, so it is not necessary to
		// decrement it before calling evm.Create.

		nonce := statedb.GetNonce(invocation.Caller)
		if nonce > 0 && params.Context.InitialDepth == 0 {
			statedb.SetNonce(invocation.Caller, nonce-1)
		}
		// we ignore returnData here because it holds the contract code that was just deployed
		var deployedContractAddress common.Address
		_, deployedContractAddress, gas, vmerr = evm.Create(sender, invocation.Input, gas, invocation.Value.ToInt())
		contractAddress = &deployedContractAddress
		// if there is an error evm.Create might not have incremented the nonce as expected,
		if statedb.GetNonce(invocation.Caller) != nonce && params.Context.InitialDepth == 0 {
			statedb.SetNonce(invocation.Caller, nonce)
		}
	} else {
		if invocation.ReadOnly {
			returnData, gas, vmerr = evm.StaticCall(sender, *invocation.Callee, invocation.Input, gas)
		} else {
			returnData, gas, vmerr = evm.Call(sender, *invocation.Callee, invocation.Input, gas, invocation.Value.ToInt())
		}
	}

	// no error means successful transaction, otherwise failure
	evmError := ""
	if vmerr != nil {
		evmError = vmerr.Error()
	}

	if returnData == nil {
		// we want [] rather than null in the response
		returnData = make([]byte, 0)
	}

	return nil, &InvocationResult{
		ReturnData:      returnData,
		LeftOverGas:     hexutil.Uint64(gas),
		ExecutionError:  evmError,
		Reverted:        vmerr == vm.ErrExecutionReverted,
		ContractAddress: contractAddress,
	}
}
