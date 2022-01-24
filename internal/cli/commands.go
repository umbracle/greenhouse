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
		UI: &cli.ColoredUi{
			ErrorColor: cli.UiColorRed,
			WarnColor:  cli.UiColorYellow,
			InfoColor:  cli.UiColorGreen,
			Ui:         ui,
		},
	}
	return map[string]cli.CommandFactory{
		"build": func() (cli.Command, error) {
			return &BuildCommand{
				baseCommand: baseCommand,
			}, nil
		},
		"clean": func() (cli.Command, error) {
			return &CleanCommand{
				baseCommand: baseCommand,
			}, nil
		},
		"test": func() (cli.Command, error) {
			return &TestCommand{
				baseCommand: baseCommand,
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
