package lib

import (
	"github.com/HorizenOfficial/go-ethereum/core/rawdb"
	"github.com/HorizenOfficial/go-ethereum/core/state"
	"github.com/HorizenOfficial/go-ethereum/ethdb"
	"github.com/HorizenOfficial/go-ethereum/log"
	"github.com/HorizenOfficial/go-ethereum/trie"
	"github.com/HorizenOfficial/go-ethereum/trie/triedb/hashdb"
)

type Database struct {
	storage  ethdb.Database
	database state.Database
}

type DatabaseParams struct {
	DatabaseHandle int `json:"databaseHandle"`
}

type LevelDBParams struct {
	Path string `json:"path"`
}

func (s *Service) open(storage ethdb.Database) int {
	db := &Database{
		storage:  storage,
		database: state.NewDatabaseWithConfig(storage, &trie.Config{HashDB: &hashdb.Config{CleanCacheSize: 256 * 1024 * 1024}}),
	}
	return s.databases.Add(db)
}

func (s *Service) DatabaseOpenMemoryDB() int {
	log.Info("initializing memorydb")
	return s.open(rawdb.NewMemoryDatabase())
}

func (s *Service) DatabaseOpenLevelDB(params LevelDBParams) (error, int) {
	log.Info("initializing leveldb", "path", params.Path)
	storage, err := rawdb.NewLevelDBDatabase(params.Path, 256, 0, "zen/db/data/", false)
	if err != nil {
		log.Error("failed to initialize database", "error", err)
		return err, 0
	}
	return nil, s.open(storage)
}

func (s *Service) DatabaseClose(params DatabaseParams) error {
	err, db := s.databases.Get(params.DatabaseHandle)
	if err != nil {
		return err
	}
	err = db.storage.Close()
	if err != nil {
		log.Error("failed to close storage", "error", err)
	}
	s.databases.Remove(params.DatabaseHandle)
	return err
}
