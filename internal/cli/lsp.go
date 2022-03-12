package cli

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mitchellh/cli"
	"github.com/umbracle/greenhouse/internal/langserver"
	"github.com/umbracle/greenhouse/internal/langserver/handlers"

	flag "github.com/spf13/pflag"
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
	var port uint64

	flag.Uint64Var(&port, "port", 0, "")
	flag.Parse()

	srv := langserver.NewLangServer(context.Background(), handlers.NewSession)
	srv.SetLogger(log.New(os.Stdout, "", 0))

	if port != 0 {
		err := srv.StartTCP(fmt.Sprintf("localhost:%d", port))
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to start TCP server: %s", err))
			return 1
		}
		return 0
	}

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
