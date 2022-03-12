package cli

import (
	flag "github.com/spf13/pflag"
)

// FmtCommand is the command to show the version of the agent
type FmtCommand struct {
	*baseCommand
}

// Help implements the cli.Command interface
func (c *FmtCommand) Help() string {
	return `Usage: greenhouse fmt

  Format contracts`
}

// Synopsis implements the cli.Command interface
func (c *FmtCommand) Synopsis() string {
	return "Format contracts"
}

func (c *FmtCommand) Flags() *flag.FlagSet {
	flags := c.baseCommand.Flags("fmt")

	return flags
}

// Run implements the cli.Command interface
func (c *FmtCommand) Run(args []string) int {
	if err := c.Flags().Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	return 0
}
