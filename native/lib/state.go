package lib

import (
	"github.com/HorizenOfficial/go-ethereum/common"
	"github.com/HorizenOfficial/go-ethereum/common/hexutil"
	"github.com/HorizenOfficial/go-ethereum/core/state"
	"github.com/HorizenOfficial/go-ethereum/core/vm"
	"github.com/HorizenOfficial/go-ethereum/crypto"
	"github.com/HorizenOfficial/go-ethereum/log"
	"os"
)

var emptyCodeHash = crypto.Keccak256Hash(nil)

type StateParams struct {
	DatabaseParams
	Root common.Hash `json:"root"`
}

type HandleParams struct {
	Handle int `json:"handle"`
}

type AccountParams struct {
	HandleParams
	Address common.Address `json:"address"`
}

type BalanceParams struct {
	AccountParams
	Amount *hexutil.Big `json:"amount"`
}

type NonceParams struct {
	AccountParams
	Nonce hexutil.Uint64 `json:"nonce"`
}

type CodeParams struct {
	AccountParams
	Code []byte `json:"code"`
}

type SnapshotParams struct {
	HandleParams
	RevisionId int `json:"revisionId"`
}

type DumpParams struct {
	HandleParams
	DumpFile string `json:"dumpFile"`
}

// StateOpen will create a new state at the given root hash.
// If the root hash is zero (or the hash of zero) this will give an empty trie.
// If the hash is anything else this will result in an error if the nodes cannot be found.
func (s *Service) StateOpen(params StateParams) (error, int) {
	err, db := s.databases.Get(params.DatabaseHandle)
	if err != nil {
		return err, 0
	}
	// TODO: research if we want to use the snapshot feature
	statedb, err := state.New(params.Root, db.database, nil)
	if err != nil {
		log.Error("failed to open state", "root", params.Root, "error", err)
		return err, 0
	}
	return nil, s.statedbs.Add(statedb)
}

func (s *Service) StateClose(params HandleParams) {
	s.statedbs.Remove(params.Handle)
}

func (s *Service) StateFinalize(params HandleParams) error {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err
	}
	statedb.Finalise(true)
	return nil
}

func (s *Service) StateIntermediateRoot(params HandleParams) (error, common.Hash) {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err, common.Hash{}
	}
	return nil, statedb.IntermediateRoot(true)
}

func (s *Service) StateCommit(params HandleParams) (error, common.Hash) {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err, common.Hash{}
	}
	//TODO ST we don't have a block number and it is not used in HashDB. It is used in PathDB, so maybe we can modify
	//this in order to support the new type of db
	hash, err := statedb.Commit(0, true)
	if err != nil {
		return err, common.Hash{}
	}
	err = statedb.Database().TrieDB().Commit(hash, false)
	if err != nil {
		return err, common.Hash{}
	}
	return nil, hash
}

// StateEmpty tests if the given account is empty,
// "empty" means non-existent or nonce==0 && balance==0 && codeHash==emptyHash (hash of nil)
func (s *Service) StateEmpty(params AccountParams) (error, bool) {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err, false
	}
	return nil, statedb.Empty(params.Address)
}

// StateIsEoa tests if a given account is an EOA or not,
// by verifying the address does not match any of the precompiled native contracts and there is no code in the account
func (s *Service) StateIsEoa(params AccountParams) (error, bool) {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err, false
	}
	// test for addresses of precompiled native contracts
	if _, ok := vm.PrecompiledContractsBerlin[params.Address]; ok {
		return nil, false
	}
	// test for code in the account
	// note: for empty accounts the code hash will be zero,
	// for existing accounts without code the hash will be the empty-hash,
	// if neither is true it must be a smart contract account
	if codeHash := statedb.GetCodeHash(params.Address); codeHash != emptyCodeHash && codeHash != (common.Hash{}) {
		return nil, false
	}
	return nil, true
}

func (s *Service) StateGetBalance(params AccountParams) (error, *hexutil.Big) {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err, nil
	}
	return nil, (*hexutil.Big)(statedb.GetBalance(params.Address))
}

func (s *Service) StateAddBalance(params BalanceParams) error {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err
	}
	statedb.AddBalance(params.Address, params.Amount.ToInt())
	return nil
}

func (s *Service) StateSubBalance(params BalanceParams) error {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err
	}
	statedb.SubBalance(params.Address, params.Amount.ToInt())
	return nil
}

func (s *Service) StateSetBalance(params BalanceParams) error {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err
	}
	statedb.SetBalance(params.Address, params.Amount.ToInt())
	return nil
}

func (s *Service) StateGetNonce(params AccountParams) (error, hexutil.Uint64) {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err, 0
	}
	return nil, (hexutil.Uint64)(statedb.GetNonce(params.Address))
}

func (s *Service) StateSetNonce(params NonceParams) error {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err
	}
	statedb.SetNonce(params.Address, uint64(params.Nonce))
	return nil
}

func (s *Service) StateGetCodeHash(params AccountParams) (error, common.Hash) {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err, common.Hash{}
	}
	return nil, statedb.GetCodeHash(params.Address)
}

func (s *Service) StateGetCode(params AccountParams) (error, []byte) {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err, nil
	}
	return nil, statedb.GetCode(params.Address)
}

// StateSetCode sets the given code, the code hash is updated automatically
func (s *Service) StateSetCode(params CodeParams) error {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err
	}
	statedb.SetCode(params.Address, params.Code)
	return nil
}

func (s *Service) StateSnapshot(params HandleParams) (error, int) {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err, 0
	}
	return nil, statedb.Snapshot()
}

func (s *Service) StateRevertToSnapshot(params SnapshotParams) error {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err
	}
	statedb.RevertToSnapshot(params.RevisionId)
	return nil
}

func (s *Service) StateDump(params DumpParams) error {
	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err
	}
	file, err := os.Create(params.DumpFile)
	if err != nil {
		return err
	}
	defer file.Close()

	opts := &state.DumpConfig{
		SkipCode:    false,
		SkipStorage: false,
	}
	_, err = file.WriteString(string(statedb.Dump(opts)))
	if err != nil {
		return err
	}

	return nil
}
