package lib

import (
	"fmt"
	"github.com/HorizenOfficial/go-ethereum/core/state"
	"github.com/HorizenOfficial/go-ethereum/eth/tracers"
	"libevm/interop"
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

var proxy CallbackProxy

func SetCallbackProxy(handler CallbackProxy) {
	proxy = handler
}

// Callback is a wrapper around an integer handle
type Callback int

// Invoke the global proxy with the handle of this callback instance
// also handles JSON serialization of arguments and deserialization of the return value
func (c *Callback) Invoke(args any, ret any) error {
	if proxy == nil {
		// TODO: is this an error case?
		return nil
	}
	argsJson, encodeErr := interop.Serialize(args)
	if encodeErr != nil {
		// note: we don't log this error because it would cause infinite recursion if we are in the log callback
		return encodeErr
	}
	result := proxy(int(*c), argsJson)
	if result == "" {
		if ret == nil {
			// 1. no result and none expected => good
			return nil
		} else {
			// 2. no result but expected one => bad
			return fmt.Errorf("callback returned nothing, but a response was expected of type: %v", ret)
		}
	}
	if ret == nil {
		// 3. got a result but none expected => bad
		return fmt.Errorf("expected empty response from callback, but got: %v", result)
	}
	// 4. got a result as expected => good, decode json
	return interop.Deserialize(result, ret)
}

// UnmarshalJSON reads a callback handle from a JSON number
func (c *Callback) UnmarshalJSON(input []byte) error {
	var handle, err = strconv.Atoi(string(input))
	if err == nil {
		*c = Callback(handle)
	}
	return err
}
