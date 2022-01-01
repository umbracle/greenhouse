package cli

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/mitchellh/cli"
)

// InitCommand is the command to show the version of the agent
type InitCommand struct {
	UI cli.Ui
}

// Help implements the cli.Command interface
func (c *InitCommand) Help() string {
	return `Usage: greenhouse version
  Display the Greenhouse version`
}

// Synopsis implements the cli.Command interface
func (c *InitCommand) Synopsis() string {
	return "Display the Greenhouse version"
}

// Run implements the cli.Command interface
func (c *InitCommand) Run(args []string) int {
	if _, err := os.Stat(defaultConfigFileName); errors.Is(err, os.ErrNotExist) {
		if err := ioutil.WriteFile(defaultConfigFileName, []byte(configTpl), 0755); err != nil {
			c.UI.Error(err.Error())
			return 1
		}
		c.UI.Output("Config file created")
	} else {
		c.UI.Output("Config file already exists")
		return 1
	}
	return 0
}

var configTpl = `
solidity = "0.8.5"
`
