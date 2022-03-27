package langserver

import (
	"fmt"

	"github.com/creachadair/jrpc2"
)

type LangServer struct {
}

func NewLangServer() *LangServer {
	opts := &jrpc2.ServerOptions{
		AllowPush:   true,
		Concurrency: 8,
	}
	fmt.Println(opts)
	srv := &LangServer{}
	return srv
}
