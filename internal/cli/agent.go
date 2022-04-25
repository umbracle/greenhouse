package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
	flag "github.com/spf13/pflag"
	"github.com/umbracle/greenhouse/internal/agent"
)

// AgentCommand is the command to show the version of the agent
type AgentCommand struct {
	*baseCommand
}

// Help implements the cli.Command interface
func (c *AgentCommand) Help() string {
	return `Usage: greenhouse agent

  Start the greenhouse agent`
}

// Synopsis implements the cli.Command interface
func (c *AgentCommand) Synopsis() string {
	return "Start the greenhouse agent"
}

func (c *AgentCommand) Flags() *flag.FlagSet {
	flags := c.baseCommand.Flags("agent")

	return flags
}

// Run implements the cli.Command interface
func (c *AgentCommand) Run(args []string) int {
	if err := c.Flags().Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Level: hclog.Info,
	})

	agent, err := agent.NewAgent(logger, nil)
	if err != nil {
		return 1
	}

	return c.handleSignals(agent.Close)
}

func (c *AgentCommand) handleSignals(closeFn func()) int {
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	sig := <-signalCh

	c.UI.Output(fmt.Sprintf("Caught signal: %v", sig))
	c.UI.Output("Gracefully shutting down agent...")

	gracefulCh := make(chan struct{})
	go func() {
		if closeFn != nil {
			closeFn()
		}
		close(gracefulCh)
	}()

	select {
	case <-signalCh:
		return 1
	case <-time.After(5 * time.Second):
		return 1
	case <-gracefulCh:
		return 0
	}
}
