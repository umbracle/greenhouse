package state

import (
	"github.com/hashicorp/go-memdb"
)

type State struct {
	db *memdb.MemDB
}

func NewState() (*State, error) {
	db, err := memdb.NewMemDB(dbSchema)
	if err != nil {
		return nil, err
	}
	s := &State{
		db: db,
	}
	return s, nil
}

func (s *State) SetTaintedSource(dir, filename string) error {
	txn := s.db.Txn(true)
	defer txn.Abort()

	obj, err := txn.First(sourcesTable, "id", dir, filename)
	if err != nil {
		return err
	}
	src := obj.(*Source)
	src = src.Copy()

	src.Tainted = true
	if err := txn.Insert(sourcesTable, src); err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *State) UpsertSource(src *Source) error {
	txn := s.db.Txn(true)
	defer txn.Abort()

	if err := txn.Insert(sourcesTable, src); err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *State) ListSources() ([]*Source, error) {
	txn := s.db.Txn(false)
	it, err := txn.Get(sourcesTable, "id")
	if err != nil {
		return nil, err
	}

	sources := make([]*Source, 0)
	for item := it.Next(); item != nil; item = it.Next() {
		source := item.(*Source)
		sources = append(sources, source)
	}

	return sources, nil
}

func (s *State) ListTaintedSources() ([]*Source, error) {
	txn := s.db.Txn(false)
	it, err := txn.Get(sourcesTable, "tainted", true)
	if err != nil {
		return nil, err
	}

	sources := make([]*Source, 0)
	for item := it.Next(); item != nil; item = it.Next() {
		source := item.(*Source)
		sources = append(sources, source)
	}

	return sources, nil
}

func (s *State) ListContracts() ([]*Contract, error) {
	txn := s.db.Txn(false)
	it, err := txn.Get(contractsTable, "id")
	if err != nil {
		return nil, err
	}

	contracts := make([]*Contract, 0)
	for item := it.Next(); item != nil; item = it.Next() {
		contract := item.(*Contract)
		contracts = append(contracts, contract)
	}

	return contracts, nil
}

func (s *State) UpsertContract(contract *Contract) error {
	txn := s.db.Txn(true)
	defer txn.Abort()

	if err := txn.Insert(contractsTable, contract); err != nil {
		return err
	}

	txn.Commit()
	return nil
}
