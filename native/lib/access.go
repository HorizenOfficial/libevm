package lib

import (
	"github.com/HorizenOfficial/go-ethereum/common"
	"github.com/HorizenOfficial/go-ethereum/core/types"
	"github.com/HorizenOfficial/go-ethereum/core/vm"
	gethparams "github.com/HorizenOfficial/go-ethereum/params"
)

type AccessParams struct {
	AccountParams
	Coinbase    common.Address   `json:"coinbase"`
	Destination *common.Address  `json:"destination"`
	AccessList  types.AccessList `json:"accessList"`
}

type SlotParams struct {
	AccountParams
	Slot common.Hash `json:"slot"`
}

func (s *Service) AccessSetup(params AccessParams) error {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err
	}
	//TODO ST For now Shanghai is not enabled but in the future we have to change here
	statedb.Prepare(gethparams.Rules{IsBerlin: true, IsShanghai: false}, params.Address, params.Coinbase,
		params.Destination, vm.PrecompiledAddressesBerlin, params.AccessList)
	return nil
}

func (s *Service) AccessAccount(params AccountParams) (error, bool) {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err, false
	}
	warmAccount := statedb.AddressInAccessList(params.Address)
	if !warmAccount {
		statedb.AddAddressToAccessList(params.Address)
	}
	return nil, warmAccount
}

func (s *Service) AccessSlot(params SlotParams) (error, bool) {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err, false
	}
	_, warmSlot := statedb.SlotInAccessList(params.Address, params.Slot)
	if !warmSlot {
		statedb.AddSlotToAccessList(params.Address, params.Slot)
	}
	return nil, warmSlot
}
