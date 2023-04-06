package lib

import (
	"encoding/json"
	"fmt"
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
