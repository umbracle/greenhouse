package cli

import (
	"fmt"

	"github.com/umbracle/greenhouse/internal/core"
	"github.com/umbracle/greenhouse/internal/lib/flagset"
)

// TestCommand is the command to show the version of the agent
type TestCommand struct {
	*baseCommand

	prefix string
}

// Help implements the cli.Command interface
func (c *TestCommand) Help() string {
	return `Usage: greenhouse test
  Test the Greenhouse project`
}

// Synopsis implements the cli.Command interface
func (c *TestCommand) Synopsis() string {
	return "Test the Greenhouse project"
}

func (c *TestCommand) Flags() *flagset.Flagset {
	flags := c.baseCommand.Flags("test")

	flags.StringFlag(&flagset.StringFlag{
		Name:  "run",
		Value: &c.prefix,
	})
	return flags
}

// Run implements the cli.Command interface
func (c *TestCommand) Run(args []string) int {
	if err := c.Flags().Parse(args); err != nil {
		panic(err)
	}

	if err := c.Init(); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	fmt.Println("a")
	input := &core.TestInput{
		Prefix: c.prefix,
	}
	if err := c.project.Test(input); err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	fmt.Println("b")

	c.UI.Output("testing")
	return 0
}
