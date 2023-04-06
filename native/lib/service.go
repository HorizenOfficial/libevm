package lib

import (
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"strconv"
)

type Service struct {
	databases *Handles[*Database]
	statedbs  *Handles[*state.StateDB]
	tracers   *Handles[*tracers.Tracer]
}

func New() *Service {
	return &Service{
		databases: NewHandles[*Database](),
		statedbs:  NewHandles[*state.StateDB](),
		tracers:   NewHandles[*tracers.Tracer](),
	}
}

type CallbackProxy func(int, string) string

var proxy CallbackProxy = func(int, string) string { return "" }

func SetCallbackProxy(handler CallbackProxy) {
	proxy = handler
}

// Callback is a wrapper around an integer handle
type Callback int

// Invoke the global proxy with the handle of this callback instance
func (c *Callback) Invoke(args string) string {
	return proxy(int(*c), args)
}

// UnmarshalJSON reads a callback handle from a JSON number
func (c *Callback) UnmarshalJSON(input []byte) error {
	var handle, err = strconv.Atoi(string(input))
	if err == nil {
		*c = Callback(handle)
	}
	return err
}
