package cli

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"
)

// GraphCommand is the command to show the version of the agent
type GraphCommand struct {
	*baseCommand

	out string

	png bool
	svg bool
}

// Help implements the cli.Command interface
func (c *GraphCommand) Help() string {
	fmt.Println(c.Flags().FlagUsages())
	return `Usage: greenhouse install
  Test the Greenhouse project`
}

// Synopsis implements the cli.Command interface
func (c *GraphCommand) Synopsis() string {
	return "Test the Greenhouse project"
}

func (c *GraphCommand) Flags() *flag.FlagSet {
	flags := c.baseCommand.Flags("graph")

	flags.StringVar(&c.out, "out", "graph", "")
	flags.BoolVar(&c.png, "png", false, "")
	flags.BoolVar(&c.svg, "svg", false, "")

	return flags
}

// Run implements the cli.Command interface
func (c *GraphCommand) Run(args []string) int {
	flags := c.Flags()
	if err := flags.Parse(args); err != nil {
		panic(err)
	}

	if err := c.Init(); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	dotGraph, err := c.project.Graph()
	if err != nil {
		panic(err)
	}

	var format string
	if c.svg {
		format = "svg"
	} else if c.png {
		format = "png"
	} else {
		// default
		format = "svg"
	}

	filename := c.out + "." + format
	if err := dotGraphToImage(dotGraph, filename); err != nil {
		panic(err)
	}

	c.UI.Output("Graph file created " + filename)
	return 0
}

func dotGraphToImage(input []byte, file string) error {
	ext := filepath.Ext(file)
	ext = strings.TrimPrefix(ext, ".")

	cmd := exec.Command("dot", "-T"+ext, "-o", file)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, string(input))
	}()

	if _, err := cmd.CombinedOutput(); err != nil {
		return err
	}
	return nil
}
