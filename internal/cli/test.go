package cli

import "fmt"

// TestCommand is the command to test the Solidity project
type TestCommand struct {
	*baseCommand
}

// Help implements the cli.Command interface
func (b *TestCommand) Help() string {
	return `Usage: greenhouse test

  Test the project`
}

// Synopsis implements the cli.Command interface
func (b *TestCommand) Synopsis() string {
	return "Test the project"
}

// Run implements the cli.Command interface
func (b *TestCommand) Run(args []string) int {
	flags := b.Flags("test")
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
	}
	return 0
}
