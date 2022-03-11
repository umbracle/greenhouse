package cli

import (
	"os"

	"github.com/mitchellh/cli"
)

func Commands() map[string]cli.CommandFactory {
	ui := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	baseCommand := &baseCommand{
		UI: ui,
	}
	return map[string]cli.CommandFactory{
		"build": func() (cli.Command, error) {
			return &BuildCommand{
				baseCommand: baseCommand,
			}, nil
		},
		"test": func() (cli.Command, error) {
			return &TestCommand{
				baseCommand: baseCommand,
			}, nil
		},
		"graph": func() (cli.Command, error) {
			return &GraphCommand{
				baseCommand: baseCommand,
			}, nil
		},
		"clean": func() (cli.Command, error) {
			return &CleanCommand{
				baseCommand: baseCommand,
			}, nil
		},
		"lsp": func() (cli.Command, error) {
			return &LspCommand{
				UI: ui,
			}, nil
		},
		"init": func() (cli.Command, error) {
			return &InitCommand{
				UI: ui,
			}, nil
		},
		"version": func() (cli.Command, error) {
			return &VersionCommand{
				UI: ui,
			}, nil
		},
	}
}
