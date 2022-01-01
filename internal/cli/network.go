package cli

import (
	"github.com/mitchellh/cli"
)

// NetworkCommand is the command to show the version of the agent
type NetworkCommand struct {
	UI cli.Ui
}

// Help implements the cli.Command interface
func (c *NetworkCommand) Help() string {
	return `Usage: greenhouse version
  Display the Greenhouse version`
}

// Synopsis implements the cli.Command interface
func (c *NetworkCommand) Synopsis() string {
	return "Display the Greenhouse version"
}

// Run implements the cli.Command interface
func (c *NetworkCommand) Run(args []string) int {
	return 0
}
