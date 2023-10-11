package lib

import (
	"github.com/HorizenOfficial/go-ethereum/common"
	"github.com/HorizenOfficial/go-ethereum/common/hexutil"
	"github.com/HorizenOfficial/go-ethereum/crypto"
	"github.com/HorizenOfficial/go-ethereum/log"
	"math/big"
)

type BlockHashCallback struct{ Callback }

func (c *BlockHashCallback) getBlockHash(blockNumber uint64) common.Hash {
	blockNumberBig := new(big.Int).SetUint64(blockNumber)
	if c == nil {
		// fallback to mocked block hash
		return common.BytesToHash(crypto.Keccak256([]byte(blockNumberBig.String())))
	}
	result := new(common.Hash)
	err := c.Invoke((*hexutil.Big)(blockNumberBig), result)
	if err != nil {
		log.Error("block hash getter callback failed: %v", err)
		return common.Hash{}
	}
	return *result
}
