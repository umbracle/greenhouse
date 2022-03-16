package cli

import (
	"fmt"
	"strings"

	flag "github.com/spf13/pflag"
)

// TestCommand is the command to test the Solidity project
type TestCommand struct {
	*baseCommand

	verbose bool
}

// Help implements the cli.Command interface
func (b *TestCommand) Help() string {
	return `Usage: greenhouse test

  Test the project

` + b.Flags().FlagUsages()
}

// Synopsis implements the cli.Command interface
func (b *TestCommand) Synopsis() string {
	return "Test the project"
}

func (b *TestCommand) Flags() *flag.FlagSet {
	flags := b.baseCommand.Flags("test")

	flags.BoolVarP(&b.verbose, "verbose", "v", false, "Show in stdout the output of the test")

	return flags
}

// Run implements the cli.Command interface
func (b *TestCommand) Run(args []string) int {
	flags := b.Flags()
	if err := flags.Parse(args); err != nil {
		b.UI.Error(err.Error())
		return 1
	}

	if err := b.Init(); err != nil {
		b.UI.Error(err.Error())
		return 1
	}

	outputs, err := b.project.Test()
	if err != nil {
		b.UI.Error(err.Error())
		return 1
	}

	for _, o := range outputs {
		res := ""
		if o.Output.Success {
			res = "[green]success[reset]"
		} else {
			res = "[red]failed[reset]"
		}
		b.UI.Output(b.Colorize().Color(fmt.Sprintf("  %s:%s:%s (%s)", o.Source, o.Contract, o.Method, res)))
		if b.verbose {
			for _, console := range o.Console {
				b.UI.Output("[" + strings.Join(console.Val, ", ") + "]")
			}
		}
	}
	return 0
}
