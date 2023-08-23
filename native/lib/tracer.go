package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/eth/tracers/logger"

	// Force-load the tracer engines to trigger registration
	_ "github.com/ethereum/go-ethereum/eth/tracers/native"
)

type TracerCreateParams struct {
	EnableMemory     bool            `json:"enableMemory"`
	DisableStack     bool            `json:"disableStack"`
	DisableStorage   bool            `json:"disableStorage"`
	EnableReturnData bool            `json:"enableReturnData"`
	Tracer           string          `json:"tracer"`
	TracerConfig     json.RawMessage `json:"tracerConfig"`
}

type TracerParams struct {
	TracerHandle int `json:"tracerHandle"`
}

type TracerResult struct {
	Result json.RawMessage `json:"result,omitempty"`
}

type TracerTxStartParams struct {
	TracerParams
	GasLimit hexutil.Uint64 `json:"gasLimit"`
}

type TracerTxEndParams struct {
	TracerParams
	RestGas hexutil.Uint64 `json:"restGas"`
}

type TracerStartParams struct {
	TracerParams
	StateDB int            `json:"stateDB"`
	Context EvmContext     `json:"context"`
	From    common.Address `json:"from"`
	To      common.Address `json:"to"`
	Create  bool           `json:"create"`
	Input   []byte         `json:"input"`
	Gas     hexutil.Uint64 `json:"gas"`
	Value   *hexutil.Big   `json:"value"`
}

type TracerEndParams struct {
	TracerParams
	Output   []byte         `json:"output"`
	GasUsed  hexutil.Uint64 `json:"gasUsed"`
	Duration time.Duration  `json:"duration"`
	Err      string         `json:"err"`
}

type TracerEnterParams struct {
	TracerParams
	OpCode string         `json:"opCode"`
	From   common.Address `json:"from"`
	To     common.Address `json:"to"`
	Input  []byte         `json:"input"`
	Gas    hexutil.Uint64 `json:"gas"`
	Value  *hexutil.Big   `json:"value"`
}

type TracerExitParams struct {
	TracerParams
	Output  []byte         `json:"output"`
	GasUsed hexutil.Uint64 `json:"gasUsed"`
	Err     string         `json:"err"`
}

func (t *TracerCreateParams) createTracer() (tracers.Tracer, error) {
	if t == nil {
		return nil, nil
	}
	if t.Tracer != "" {
		tracer, err := tracers.New(t.Tracer, nil, t.TracerConfig)
		if err != nil {
			return nil, err
		}
		return tracer, nil
	} else {
		// default to the struct logger
		traceConfig := logger.Config{
			EnableMemory:     t.EnableMemory,
			DisableStack:     t.DisableStack,
			DisableStorage:   t.DisableStorage,
			EnableReturnData: t.EnableReturnData,
		}
		return logger.NewStructLogger(&traceConfig), nil
	}
}

func (s *Service) TracerCreate(params TracerCreateParams) (error, int) {
	tracer, err := params.createTracer()
	if err != nil {
		return err, 0
	}
	return nil, s.tracers.Add(&tracer)
}

func (s *Service) TracerRemove(params TracerParams) {
	s.tracers.Remove(params.TracerHandle)
}

func (s *Service) TracerResult(params TracerParams) (error, *TracerResult) {
	err, tracerPtr := s.tracers.Get(params.TracerHandle)
	if err != nil {
		return err, nil
	}
	tracer := *tracerPtr
	traceResultJson, err := tracer.GetResult()
	if err != nil {
		return fmt.Errorf("trace error: %v", err), nil
	}
	return nil, &TracerResult{Result: traceResultJson}
}

// TracerCaptureTxStart maps to CaptureTxStart(gasLimit uint64)
func (s *Service) TracerCaptureTxStart(params TracerTxStartParams) error {
	err, tracerPtr := s.tracers.Get(params.TracerHandle)
	if err != nil {
		return err
	}
	tracer := *tracerPtr
	tracer.CaptureTxStart(uint64(params.GasLimit))
	return nil
}

// TracerCaptureTxEnd maps to CaptureTxEnd(restGas uint64)
func (s *Service) TracerCaptureTxEnd(params TracerTxEndParams) error {
	err, tracerPtr := s.tracers.Get(params.TracerHandle)
	if err != nil {
		return err
	}
	tracer := *tracerPtr
	tracer.CaptureTxEnd(uint64(params.RestGas))
	return nil
}

// TracerCaptureStart maps to CaptureStart(env *EVM, from common.Address, to common.Address, create bool, input []byte, gas uint64, value *big.Int)
func (s *Service) TracerCaptureStart(params TracerStartParams) error {
	err, tracerPtr := s.tracers.Get(params.TracerHandle)
	if err != nil {
		return err
	}
	tracer := *tracerPtr

	err, stateDB := s.statedbs.Get(params.StateDB)
	if err != nil {
		return err
	}

	err, evm := s.getEvm(params.Context, stateDB, params.From)
	if err != nil {
		return err
	}

	tracer.CaptureStart(
		evm,
		params.From,
		params.To,
		params.Create,
		params.Input,
		uint64(params.Gas),
		params.Value.ToInt(),
	)
	return nil
}

// TracerCaptureEnd maps to CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error)
func (s *Service) TracerCaptureEnd(params TracerEndParams) error {
	err, tracerPtr := s.tracers.Get(params.TracerHandle)
	if err != nil {
		return err
	}
	tracer := *tracerPtr
	var traceErr error
	if params.Err != "" {
		traceErr = errors.New(params.Err)
	}
	tracer.CaptureEnd(params.Output, uint64(params.GasUsed), params.Duration, traceErr)
	return nil
}

// TracerCaptureEnter maps to CaptureEnter(typ OpCode, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int)
func (s *Service) TracerCaptureEnter(params TracerEnterParams) error {
	err, tracerPtr := s.tracers.Get(params.TracerHandle)
	if err != nil {
		return err
	}

	value := params.Value.ToInt()
	//In Geth in case of STATICCALL value is nil, so it doesn't appear in the tracer result
	if params.OpCode == "STATICCALL" {
		value = nil
	}

	tracer := *tracerPtr
	tracer.CaptureEnter(
		vm.StringToOp(params.OpCode),
		params.From,
		params.To,
		params.Input,
		uint64(params.Gas),
		value,
	)
	return nil
}

// TracerCaptureExit maps to CaptureExit(output []byte, gasUsed uint64, err error)
func (s *Service) TracerCaptureExit(params TracerExitParams) error {
	err, tracerPtr := s.tracers.Get(params.TracerHandle)
	if err != nil {
		return err
	}
	tracer := *tracerPtr
	var traceErr error
	if params.Err != "" {
		traceErr = errors.New(params.Err)
	}
	tracer.CaptureExit(params.Output, uint64(params.GasUsed), traceErr)
	return nil
}
