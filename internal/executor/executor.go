package state

import (
	"github.com/ethereum/evmc/v10/bindings/go/evmc"
	"github.com/hashicorp/go-memdb"
)

var _ evmc.HostContext = &txn{}

type Executor struct {
	db *memdb.MemDB
}

func (e *Executor) do() {
	xx := e.db.Txn(true)
	xx.Commit()
}

type txn struct {
	txn *memdb.Txn
}

func (s *txn) AccountExists(addr evmc.Address) bool {
	return false
}

func (s *txn) GetStorage(addr evmc.Address, key evmc.Hash) evmc.Hash {
	return evmc.Hash{}
}

func (s *txn) SetStorage(addr evmc.Address, key evmc.Hash, value evmc.Hash) evmc.StorageStatus {
	return 1
}

func (s *txn) GetBalance(addr evmc.Address) evmc.Hash {
	obj, err := s.txn.First("account", "address", addr)
	if err != nil {
		return evmc.Hash{}
	}
	return obj.(*Account).Balance
}

func (s *txn) GetCodeSize(addr evmc.Address) int {
	return 0
}

func (s *txn) GetCodeHash(addr evmc.Address) evmc.Hash {
	return evmc.Hash{}
}

func (s *txn) GetCode(addr evmc.Address) []byte {
	return nil
}

func (s *txn) Selfdestruct(addr evmc.Address, beneficiary evmc.Address) {
}

func (s *txn) GetTxContext() evmc.TxContext {
	return evmc.TxContext{}
}

func (s *txn) GetBlockHash(number int64) evmc.Hash {
	return evmc.Hash{}
}

func (s *txn) EmitLog(addr evmc.Address, topics []evmc.Hash, data []byte) {
}

func (s *txn) Call(kind evmc.CallKind,
	recipient evmc.Address, sender evmc.Address, value evmc.Hash, input []byte, gas int64, depth int,
	static bool, salt evmc.Hash, codeAddress evmc.Address) (output []byte, gasLeft int64, createAddr evmc.Address, err error) {

	return nil, 0, evmc.Address{}, nil
}

func (s *txn) AccessAccount(addr evmc.Address) evmc.AccessStatus {
	return evmc.ColdAccess
}

func (s *txn) AccessStorage(addr evmc.Address, key evmc.Hash) evmc.AccessStatus {
	return evmc.ColdAccess
}
