package agent

import (
	"github.com/hashicorp/go-hclog"
	"github.com/umbracle/greenhouse/internal/agent/langserver"
)

type Agent struct {
	logger hclog.Logger
	config *Config

	lsp *langserver.LangServer
}

func NewAgent(logger hclog.Logger, config *Config) (*Agent, error) {
	a := &Agent{
		logger: logger,
		config: config,
	}

	if err := a.startLangServer(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *Agent) startLangServer() error {
	a.lsp = langserver.NewLangServer(a.logger.Named("lsp"))
	if err := a.lsp.StartTCP("localhost:4564"); err != nil {
		return err
	}

	return nil
}

func (a *Agent) Close() {
	a.lsp.Close()
}
