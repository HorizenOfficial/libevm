package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
)

// SerializableConfig mirrors configurable parts of the runtime.Config
type SerializableConfig struct {
	Difficulty  BigInt         `json:"difficulty"`
	Origin      common.Address `json:"origin"`
	Coinbase    common.Address `json:"coinbase"`
	BlockNumber BigInt         `json:"blockNumber"`
	Time        BigInt         `json:"time"`
	GasLimit    uint64         `json:"gasLimit"`
	GasPrice    BigInt         `json:"gasPrice"`
	Value       BigInt         `json:"value"`
	BaseFee     BigInt         `json:"baseFee"`
}

// GetConfig maps SerializableConfig to runtime.Config
func (c *SerializableConfig) GetConfig() *runtime.Config {
	return &runtime.Config{
		ChainConfig: nil,
		Difficulty:  c.Difficulty.Int,
		Origin:      c.Origin,
		Coinbase:    c.Coinbase,
		BlockNumber: c.BlockNumber.Int,
		Time:        c.Time.Int,
		GasLimit:    c.GasLimit,
		GasPrice:    c.GasPrice.Int,
		Value:       c.Value.Int,
		BaseFee:     c.BaseFee.Int,
	}
}