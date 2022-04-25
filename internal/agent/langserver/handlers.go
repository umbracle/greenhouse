package langserver

import (
	"context"
	"fmt"

	"github.com/umbracle/greenhouse/internal/agent/langserver/lsp"
	"github.com/umbracle/greenhouse/internal/core"
)

func (l *LangServer) Initialize(ctx context.Context, params lsp.InitializeParams) (lsp.InitializeResult, error) {
	fmt.Println("root path", params.RootPath)

	config := core.DefaultConfig()
	config.DataDir = params.RootPath

	project, err := core.NewProject(l.logger, config)
	if err != nil {
		panic(err)
	}
	l.project = project

	if err := l.project.Compile(); err != nil {
		panic(err)
	}

	resp := lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: lsp.TextDocumentSyncOptions{
				OpenClose: true,
				Change:    lsp.Incremental,
			},
			CompletionProvider: lsp.CompletionOptions{
				ResolveProvider:   false,
				TriggerCharacters: []string{".", "["},
			},
			CodeActionProvider: lsp.CodeActionOptions{
				ResolveProvider: false,
			},
			DeclarationProvider:        lsp.DeclarationOptions{},
			DefinitionProvider:         true,
			CodeLensProvider:           lsp.CodeLensOptions{},
			ReferencesProvider:         true,
			HoverProvider:              true,
			DocumentFormattingProvider: true,
			DocumentSymbolProvider:     true,
		},
	}
	return resp, nil
}

func (l *LangServer) Initialized(ctx context.Context, params lsp.InitializedParams) error {
	return nil
}

func (l *LangServer) TextDocumentDidChange(ctx context.Context, params lsp.DidChangeTextDocumentParams) error {
	fmt.Println("did change -- uri --", params.TextDocument.URI)

	return nil
}

func (l *LangServer) TextDocumentDidOpen(ctx context.Context, params lsp.DidOpenTextDocumentParams) error {
	fmt.Println("did open -- uri --", params.TextDocument.URI)

	return nil
}
