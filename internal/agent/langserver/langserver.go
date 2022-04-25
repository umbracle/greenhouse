package langserver

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/creachadair/jrpc2"
	"github.com/creachadair/jrpc2/channel"
	"github.com/creachadair/jrpc2/code"
	rpch "github.com/creachadair/jrpc2/handler"
	"github.com/creachadair/jrpc2/server"
	"github.com/hashicorp/go-hclog"
	"github.com/umbracle/greenhouse/internal/core"
)

type LangServer struct {
	logger  hclog.Logger
	lis     net.Listener
	project *core.Project
}

func NewLangServer(logger hclog.Logger) *LangServer {
	opts := &jrpc2.ServerOptions{
		AllowPush:   true,
		Concurrency: 8,
	}
	fmt.Println(opts)
	srv := &LangServer{
		logger: logger,
	}
	return srv
}

func (l *LangServer) StartTCP(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("TCP Server failed to start: %s", err)
	}
	l.lis = lis

	accepter := server.NetAccepter(lis, channel.LSP)
	newService := func() server.Service {
		return l
	}

	opts := &jrpc2.ServerOptions{
		Logger: jrpc2.StdLogger(l.logger.StandardLogger(&hclog.StandardLoggerOptions{})),
		RPCLog: &rpcLogger{logger: l.logger.Named("rpc")},
	}

	go func() {
		l.logger.Info("Starting loop server ...")
		err = server.Loop(context.TODO(), accepter, newService, &server.LoopOptions{ServerOptions: opts})
		if err != nil {
			l.logger.Error("Loop server failed to start: %s", err)
		}
	}()

	return nil
}

func (l *LangServer) Assigner() (jrpc2.Assigner, error) {

	m := map[string]rpch.Func{
		"initialize": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			return handle(ctx, req, l.Initialize)
		},
		"initialized": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			return handle(ctx, req, l.Initialized)
		},
		"textDocument/codeAction": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			return nil, nil
		},
		"textDocument/codeLens": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			return nil, nil
		},
		"textDocument/definition": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			return nil, nil
		},
		"textDocument/formatting": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			return nil, nil
		},
		"textDocument/documentLink": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			return nil, nil
		},
		"textDocument/documentSymbol": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			return nil, nil
		},
		"textDocument/didChange": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			return handle(ctx, req, l.TextDocumentDidChange)
		},
		"textDocument/didOpen": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			return handle(ctx, req, l.TextDocumentDidOpen)
		},
		"textDocument/hover": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			return nil, nil
		},
	}

	return convertMap(m), nil
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

func convertMap(m map[string]rpch.Func) rpch.Map {
	hm := make(rpch.Map, len(m))

	for method, fun := range m {
		hm[method] = rpch.New(fun)
	}

	return hm
}

func (l *LangServer) Finish(jrpc2.Assigner, jrpc2.ServerStatus) {

}

func (l *LangServer) Close() {
	if err := l.lis.Close(); err != nil {
		l.logger.Error(err.Error())
	}
}
