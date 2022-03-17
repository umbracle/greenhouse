package core

import (
	"encoding/json"

	"github.com/umbracle/greenhouse/internal/state"
)

type metadataFormat struct {
	Sources   []*state.Source
	Contracts []*state.Contract
}

func getMetadataRaw(s *state.State) ([]byte, error) {
	sources, err := s.ListSources()
	if err != nil {
		return nil, err
	}

	contracts, err := s.ListContracts()
	if err != nil {
		return nil, err
	}
	out := &metadataFormat{
		Contracts: contracts,
		Sources:   sources,
	}
	data, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	return data, nil
}
