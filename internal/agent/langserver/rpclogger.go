package langserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/creachadair/jrpc2"
	"github.com/hashicorp/go-hclog"
)

type rpcLogger struct {
	logger hclog.Logger
}

func (rl *rpcLogger) LogRequest(ctx context.Context, req *jrpc2.Request) {
	idStr := ""
	if req.ID() != "" {
		idStr = fmt.Sprintf(" (ID %s)", req.ID())
	}
	reqType := "request"
	if req.IsNotification() {
		reqType = "notification"
	}

	var params json.RawMessage
	req.UnmarshalParams(&params)

	rl.logger.Trace("Request", "reqType", reqType, "method", req.Method(), "id", idStr, "params", params)
}

func (rl *rpcLogger) LogResponse(ctx context.Context, rsp *jrpc2.Response) {
	idStr := ""
	if rsp.ID() != "" {
		idStr = fmt.Sprintf(" (ID %s)", rsp.ID())
	}

	req := jrpc2.InboundRequest(ctx)
	if req.IsNotification() {
		idStr = " (notification)"
	}

	if rsp.Error() != nil {
		rl.logger.Trace("Response error", "method", req.Method(), "id", idStr, "error", rsp.Error())
		return
	}
	var body json.RawMessage
	rsp.UnmarshalResult(&body)

	rl.logger.Trace("Response", "method", req.Method(), "id", idStr, "body", body)
}
