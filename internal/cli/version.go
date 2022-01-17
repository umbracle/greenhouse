package cli

import (
	"github.com/mitchellh/cli"
	"github.com/umbracle/greenhouse/version"
)

// VersionCommand is the command to show the version of the agent
type VersionCommand struct {
	UI cli.Ui
}

// Help implements the cli.Command interface
func (c *VersionCommand) Help() string {
	return `Usage: greenhouse version
	
  Display the Greenhouse version`
}

// Synopsis implements the cli.Command interface
func (c *VersionCommand) Synopsis() string {
	return "Display the Greenhouse version"
}

// Run implements the cli.Command interface
func (c *VersionCommand) Run(args []string) int {
	c.UI.Output(version.GetVersion())
	return 0
}
