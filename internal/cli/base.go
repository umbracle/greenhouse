package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/cli"
	"github.com/mitchellh/colorstring"
	flag "github.com/spf13/pflag"
	"github.com/umbracle/greenhouse/internal/core"
)

var defaultConfigFileName = "greenhouse.hcl"

type baseCommand struct {
	UI cli.Ui

	project   *core.Project
	cliConfig *core.Config
}

func (b *baseCommand) Flags(name string) *flag.FlagSet {
	b.cliConfig = &core.Config{}

	flags := flag.NewFlagSet(name, 0)
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

func (b *baseCommand) Colorize() *colorstring.Colorize {
	_, coloredUi := b.UI.(*cli.ColoredUi)

	return &colorstring.Colorize{
		Colors:  colorstring.DefaultColors,
		Disable: !coloredUi,
		Reset:   true,
	}
}
