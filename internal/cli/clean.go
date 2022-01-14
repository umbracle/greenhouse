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
	return `Usage: greenhouse test
  Test the Greenhouse project`
}

// Synopsis implements the cli.Command interface
func (c *CleanCommand) Synopsis() string {
	return "Test the Greenhouse project"
}

func (c *CleanCommand) Flags() *flag.FlagSet {
	flags := c.baseCommand.Flags("clean")

	return flags
}

// Run implements the cli.Command interface
func (c *CleanCommand) Run(args []string) int {
	if err := c.Flags().Parse(args); err != nil {
		panic(err)
	}

	if err := os.RemoveAll(".greenhouse"); err != nil {
		panic(err)
	}

	c.UI.Output("testing")
	return 0
}
