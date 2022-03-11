package cli

import (
	"context"
	"os"

	"github.com/mitchellh/cli"
	"github.com/umbracle/greenhouse/internal/langserver"
	"github.com/umbracle/greenhouse/internal/langserver/handlers"
)

// LspCommand is the command to show the version of the agent
type LspCommand struct {
	UI cli.Ui
}

// Help implements the cli.Command interface
func (c *LspCommand) Help() string {
	return `Usage: greenhouse version
  Display the Greenhouse version`
}

// Synopsis implements the cli.Command interface
func (c *LspCommand) Synopsis() string {
	return "Display the Greenhouse version"
}

// Run implements the cli.Command interface
func (c *LspCommand) Run(args []string) int {

	srv := langserver.NewLangServer(context.Background(), handlers.NewSession)

	err := srv.StartAndWait(os.Stdin, os.Stdout)
	if err != nil {
		panic(err)
	}

	/*
		r, err := os.Open("file.txt")
		if err != nil {
			panic(err)
		}
		if _, err := io.Copy(r, os.Stdin); err != nil {
			panic(err)
		}
	*/

	return 0
}
