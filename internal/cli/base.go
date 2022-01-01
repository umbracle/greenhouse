package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/cli"
	"github.com/umbracle/greenhouse/internal/core"
	"github.com/umbracle/greenhouse/internal/lib/flagset"
)

var defaultConfigFileName = "greenhouse.hcl"

type baseCommand struct {
	UI cli.Ui

	project   *core.Project
	cliConfig *core.Config
}

func (b *baseCommand) Flags(name string) *flagset.Flagset {
	b.cliConfig = &core.Config{}

	flags := flagset.NewFlagSet(name)

	return flags
}

func (b *baseCommand) Init() error {
	config := core.DefaultConfig()

	if _, err := os.Stat(defaultConfigFileName); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("config file does not exists")
	}
	fileConfig, err := core.LoadConfig(defaultConfigFileName)
	if err != nil {
		return err
	}

	if err := config.Merge(fileConfig); err != nil {
		return err
	}

	p, err := core.NewProject(hclog.L(), config)
	if err != nil {
		return err
	}
	b.project = p
	return nil
}
