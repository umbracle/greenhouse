package state

import (
	"github.com/hashicorp/go-memdb"
)

var (
	sourcesTable   = "sources"
	contractsTable = "contracts"
)

var dbSchema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		sourcesTable: {
			Name: sourcesTable,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:   "id",
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "Dir"},
							&memdb.StringFieldIndex{Field: "Filename"},
						},
					},
				},
				"tainted": {
					Name:    "tainted",
					Indexer: &memdb.BoolFieldIndex{Field: "Tainted"},
				},
			},
		},
		contractsTable: {
			Name: contractsTable,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:   "id",
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "Name"},
						},
					},
				},
			},
		},
	},
}
