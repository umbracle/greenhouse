package handlers

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/creachadair/jrpc2"
	"github.com/creachadair/jrpc2/code"
	rpch "github.com/creachadair/jrpc2/handler"
	"github.com/umbracle/greenhouse/internal/langserver/filesystem"
	ilsp "github.com/umbracle/greenhouse/internal/langserver/lsp"
	lsp "github.com/umbracle/greenhouse/internal/langserver/protocol"
	"github.com/umbracle/greenhouse/internal/langserver/session"
)

type service struct {
	logger *log.Logger

	srvCtx     context.Context
	filesystem filesystem.Filesystem

	sessCtx     context.Context
	stopSession context.CancelFunc
	server      session.Server

	additionalHandlers map[string]rpch.Func
}

var discardLogs = log.New(ioutil.Discard, "", 0)

func NewSession(srvCtx context.Context) session.Session {

	sessCtx, stopSession := context.WithCancel(srvCtx)
	return &service{
		logger:      discardLogs,
		srvCtx:      srvCtx,
		sessCtx:     sessCtx,
		stopSession: stopSession,
	}
}

func (svc *service) SetLogger(logger *log.Logger) {
	svc.logger = logger
}

// Assigner builds out the jrpc2.Map according to the LSP protocol
// and passes related dependencies to handlers via context
func (svc *service) Assigner() (jrpc2.Assigner, error) {
	svc.logger.Println("Preparing new session ...")

	session := session.NewSession(svc.stopSession)

	err := session.Prepare()
	if err != nil {
		return nil, fmt.Errorf("Unable to prepare session: %w", err)
	}

	lh := LogHandler(svc.logger)

	cc := &lsp.ClientCapabilities{}

	m := map[string]rpch.Func{
		"initialize": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.Initialize(req)
			if err != nil {
				return nil, err
			}

			ctx = ilsp.WithClientCapabilities(ctx, cc)

			return handle(ctx, req, svc.Initialize)
		},
		"initialized": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.ConfirmInitialization(req)
			if err != nil {
				return nil, err
			}

			return handle(ctx, req, Initialized)
		},
		"textDocument/codeAction": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			return handle(ctx, req, lh.TextDocumentCodeAction)
		},
		"textDocument/codeLens": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			/*
				ctx = ilsp.WithClientCapabilities(ctx, cc)
				ctx = lsctx.WithDocumentStorage(ctx, svc.fs)
			*/
			return handle(ctx, req, svc.TextDocumentCodeLens)
		},
		/*
			"initialized": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
				err := session.ConfirmInitialization(req)
				if err != nil {
					return nil, err
				}

				return handle(ctx, req, Initialized)
			},
		*/
		"textDocument/hover": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			return handle(ctx, req, svc.TextDocumentHover)
		},
	}

	// For use in tests, e.g. to test request cancellation
	if len(svc.additionalHandlers) > 0 {
		for methodName, handlerFunc := range svc.additionalHandlers {
			m[methodName] = handlerFunc
		}
	}

	return convertMap(m), nil
}

func Initialized(ctx context.Context, params lsp.InitializedParams) error {
	return nil
}

func (svc *service) TextDocumentHover(ctx context.Context, params lsp.TextDocumentPositionParams) (*lsp.Hover, error) {
	fmt.Println("// logger //")
	return &lsp.Hover{Contents: lsp.MarkupContent{Kind: lsp.PlainText, Value: "abcd"}, Range: lsp.Range{Start: lsp.Position{Line: 10, Character: 20}, End: lsp.Position{Line: 11, Character: 20}}}, nil
}

func (svc *service) Finish(_ jrpc2.Assigner, status jrpc2.ServerStatus) {
	if status.Closed || status.Err != nil {
		svc.logger.Printf("session stopped unexpectedly (err: %v)", status.Err)
	}

	svc.shutdown()
	svc.stopSession()
}

func (svc *service) shutdown() {
}

// convertMap is a helper function allowing us to omit the jrpc2.Func
// signature from the method definitions
func convertMap(m map[string]rpch.Func) rpch.Map {
	hm := make(rpch.Map, len(m))

	for method, fun := range m {
		hm[method] = rpch.New(fun)
	}

	return hm
}

const requestCancelled code.Code = -32800

// handle calls a jrpc2.Func compatible function
func handle(ctx context.Context, req *jrpc2.Request, fn interface{}) (interface{}, error) {
	f := rpch.New(fn)
	result, err := f.Handle(ctx, req)
	if ctx.Err() != nil && errors.Is(ctx.Err(), context.Canceled) {
		err = fmt.Errorf("%w: %s", requestCancelled.Err(), err)
	}
	return result, err
}

// logHandler provides handlers logger
type logHandler struct {
	logger *log.Logger
}

func LogHandler(logger *log.Logger) *logHandler {
	return &logHandler{logger}
}

// ....

func (h *logHandler) TextDocumentCodeAction(ctx context.Context, params lsp.CodeActionParams) []lsp.CodeAction {
	ca, err := h.textDocumentCodeAction(ctx, params)
	if err != nil {
		h.logger.Printf("code action failed: %s", err)
	}

	return ca
}

func (h *logHandler) textDocumentCodeAction(ctx context.Context, params lsp.CodeActionParams) ([]lsp.CodeAction, error) {
	var ca []lsp.CodeAction

	// For action definitions, refer to https://code.visualstudio.com/api/references/vscode-api#CodeActionKind
	// We only support format type code actions at the moment, and do not want to format without the client asking for
	// them, so exit early here if nothing is requested.
	if len(params.Context.Only) == 0 {
		h.logger.Printf("No code action requested, exiting")
		return ca, nil
	}

	return ca, nil
}

func (svc *service) TextDocumentCodeLens(ctx context.Context, params lsp.CodeLensParams) ([]lsp.CodeLens, error) {
	list := make([]lsp.CodeLens, 0)

	list = append(list, lsp.CodeLens{
		Range: lsp.Range{Start: lsp.Position{Line: 10, Character: 20}, End: lsp.Position{Line: 11, Character: 20}},
		Command: lsp.Command{
			Title:   "Command1",
			Command: "command2",
		},
	})

	/*
		fs, err := lsctx.DocumentStorage(ctx)
		if err != nil {
			return list, err
		}

		fh := ilsp.FileHandlerFromDocumentURI(params.TextDocument.URI)
		doc, err := fs.GetDocument(fh)
		if err != nil {
			return list, err
		}

		path := lang.Path{
			Path:       doc.Dir(),
			LanguageID: doc.LanguageID(),
		}

		lenses, err := svc.decoder.CodeLensesForFile(ctx, path, doc.Filename())
		if err != nil {
			return nil, err
		}

		for _, lens := range lenses {
			cmd, err := ilsp.Command(lens.Command)
			if err != nil {
				svc.logger.Printf("skipping code lens %#v: %s", lens.Command, err)
				continue
			}

			list = append(list, lsp.CodeLens{
				Range:   ilsp.HCLRangeToLSP(lens.Range),
				Command: cmd,
			})
		}
	*/

	return list, nil
}
