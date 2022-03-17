package core

import (
	"encoding/json"
	"io/ioutil"

	"github.com/umbracle/greenhouse/internal/state"
)

type metadataFormat struct {
	Sources   []*state.Source
	Contracts []*state.Contract
}

func writeMetadata(s *state.State) []byte {
	sources, err := s.ListSources()
	if err != nil {
		panic(err)
	}

	contracts, err := s.ListContracts()
	if err != nil {
		panic(err)
	}
	out := &metadataFormat{
		Contracts: contracts,
		Sources:   sources,
	}
	data, err := json.Marshal(out)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile("./metadata2.json", data, 0755); err != nil {
		panic(err)
	}
	return data
}
