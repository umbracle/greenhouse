package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/imdario/mergo"
)

// Config is the greenhouse project configuration
type Config struct {
	Contracts    string
	Solidity     string
	Dependencies map[string]string
	DataDir      string
}

func DefaultConfig() *Config {
	return &Config{
		Contracts:    "contracts",
		Solidity:     "0.8.4",
		Dependencies: map[string]string{},
		DataDir:      "",
	}
}

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var f func([]byte, interface{}) error
	switch {
	case strings.HasSuffix(path, ".hcl"):
		f = hcl.Unmarshal
	case strings.HasSuffix(path, ".json"):
		f = json.Unmarshal
	default:
		return nil, fmt.Errorf("suffix of %s is neither hcl nor json", path)
	}

	var config Config
	if err := f(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Merge(cc ...*Config) error {
	for _, elem := range cc {
		if err := mergo.Merge(c, elem, mergo.WithOverride, mergo.WithAppendSlice); err != nil {
			return fmt.Errorf("failed to merge configurations: %v", err)
		}
	}
	return nil
}
