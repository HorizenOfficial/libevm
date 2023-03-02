package lib

import (
	"errors"
	"fmt"
	"math"
	"sync"
)

var ErrInvalidHandle = errors.New("invalid handle")

type Handles[T comparable] struct {
	used    map[int]T
	current int
	mutex   *sync.RWMutex
}

func NewHandles[T comparable]() *Handles[T] {
	return &Handles[T]{
		used:    make(map[int]T),
		current: 0,
		mutex:   &sync.RWMutex{},
	}
}

func (h *Handles[T]) Add(obj T) int {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	// pedantic safety check for overflow: this is highly unlikely because there are 2**31-1 available handles,
	// just the memory consumption alone will likely cause problems before this happens
	if len(h.used) == math.MaxInt32 {
		panic(fmt.Sprintf("out of handles, unable to add %T", obj))
	}
	for {
		// wrap around
		if h.current == math.MaxInt32 {
			h.current = 0
		}
		// find the next unused handle:
		// this will never give a handle of 0, which is on purpose - we might consider a handle of 0 as invalid
		h.current++
		if _, exists := h.used[h.current]; !exists {
			h.used[h.current] = obj
			return h.current
		}
	}
}

func (h *Handles[T]) Get(handle int) (error, T) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	if obj, exists := h.used[handle]; exists {
		return nil, obj
	} else {
		// this gives the default value of type T, i.e. 0 for numbers, false for bool, nil for pointer types, maps, etc.
		empty := *new(T)
		return fmt.Errorf("%w: %d", ErrInvalidHandle, handle), empty
	}
}

func (h *Handles[T]) Remove(handle int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	delete(h.used, handle)
}
