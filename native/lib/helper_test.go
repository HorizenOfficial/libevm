package lib

import (
	"github.com/HorizenOfficial/go-ethereum/common"
)

func SetupTest() (*Service, int, int) {
	var (
		instance       = New()
		dbHandle       = instance.DatabaseOpenMemoryDB()
		_, stateHandle = instance.StateOpen(StateParams{
			DatabaseParams: DatabaseParams{DatabaseHandle: dbHandle},
			Root:           common.Hash{},
		})
	)
	return instance, dbHandle, stateHandle
}
