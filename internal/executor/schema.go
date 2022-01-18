package state

import (
	"github.com/ethereum/evmc/v10/bindings/go/evmc"
	"github.com/hashicorp/go-memdb"
)

var dbSchema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		"account": {
			Name: "account",
			Indexes: map[string]*memdb.IndexSchema{
				"address": {
					Name:    "address",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "Email"},
				},
			},
		},
	},
}

type Account struct {
	Address evmc.Address
	Balance evmc.Hash
}
