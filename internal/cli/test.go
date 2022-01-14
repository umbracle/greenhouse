package cli

import (
	"fmt"

	flag "github.com/spf13/pflag"
	"github.com/umbracle/greenhouse/internal/core"
)

// TestCommand is the command to show the version of the agent
type TestCommand struct {
	*baseCommand

	verbose bool
	prefix  string
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

func (c *TestCommand) Flags() *flag.FlagSet {
	flags := c.baseCommand.Flags("test")

	flags.StringVar(&c.prefix, "run", "", "")
	flags.BoolVarP(&c.verbose, "verbose", "v", false, "")

	return flags
}

// Run implements the cli.Command interface
func (c *TestCommand) Run(args []string) int {

	flags := c.Flags()
	if err := flags.Parse(args); err != nil {
		panic(err)
	}
	args = flags.Args()

	if err := c.Init(); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	fmt.Println("a")
	input := &core.TestInput{
		Prefix:  c.prefix,
		Verbose: c.verbose,
	}
	if len(args) == 1 {
		input.Path = args[0]
	}

	if err := c.project.Test(input); err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	fmt.Println("b")

	c.UI.Output("testing")
	return 0
}
