package cli

import (
	"os"

	flag "github.com/spf13/pflag"
)

// CleanCommand is the command to show the version of the agent
type CleanCommand struct {
	*baseCommand
}

// Help implements the cli.Command interface
func (c *CleanCommand) Help() string {
	return `Usage: greenhouse clean

  Clean and reset the project`
}

// Synopsis implements the cli.Command interface
func (c *CleanCommand) Synopsis() string {
	return "Clean and reset the project"
}

func (c *CleanCommand) Flags() *flag.FlagSet {
	flags := c.baseCommand.Flags("clean")

	return flags
}

// Run implements the cli.Command interface
func (c *CleanCommand) Run(args []string) int {
	if err := c.Flags().Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	if err := os.RemoveAll(".greenhouse"); err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	return 0
}
