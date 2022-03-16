package cli

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/google/go-dap"
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

	go otherServer()

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

func otherServer() {
	lst, err := net.Listen("tcp", "localhost:4569")
	if err != nil {
		panic(err)
	}

	fmt.Println("localhost:4569")
	fmt.Println(lst.Addr())

	for {
		conn, err := lst.Accept()
		if err != nil {
			panic(err)
		}

		fmt.Println("-- conn --")
		fmt.Println(conn)

		go func() {
			defer conn.Close()
			reader := bufio.NewReader(conn)

			send := func(message dap.Message) {
				if err := dap.WriteProtocolMessage(conn, message); err != nil {
					panic(err)
				}
			}

			for {
				request, err := dap.ReadProtocolMessage(reader)
				if err != nil {
					panic(err)
				}
				fmt.Println("====================================> request --")
				fmt.Println(request)

				switch request := request.(type) {
				case *dap.InitializeRequest: // Required
					fmt.Println("A")
					fmt.Println(request)

					response := &dap.InitializeResponse{Response: *newResponse(request.Request)}
					response.Body.SupportsConfigurationDoneRequest = true
					response.Body.SupportsConditionalBreakpoints = true
					response.Body.SupportsDelayedStackTraceLoading = true
					response.Body.SupportTerminateDebuggee = true
					response.Body.SupportsFunctionBreakpoints = true
					response.Body.SupportsInstructionBreakpoints = true
					response.Body.SupportsExceptionInfoRequest = true
					response.Body.SupportsSetVariable = true
					response.Body.SupportsEvaluateForHovers = true
					response.Body.SupportsClipboardContext = true
					response.Body.SupportsSteppingGranularity = true
					response.Body.SupportsLogPoints = true
					response.Body.SupportsDisassembleRequest = true
					// TODO(polina): support these requests in addition to vscode-go feature parity
					response.Body.SupportsTerminateRequest = false
					response.Body.SupportsRestartRequest = false
					response.Body.SupportsStepBack = false // To be enabled by CapabilitiesEvent based on configuration
					response.Body.SupportsSetExpression = false
					response.Body.SupportsLoadedSourcesRequest = false
					response.Body.SupportsReadMemoryRequest = false
					response.Body.SupportsCancelRequest = false

					send(response)
					return
				case *dap.LaunchRequest: // Required
					fmt.Println("B")

					send(&dap.InitializedEvent{Event: *newEvent("initialized")})
					send(&dap.LaunchResponse{Response: *newResponse(request.Request)})
					return
				case *dap.AttachRequest: // Required
					fmt.Println("C")

					send(&dap.InitializedEvent{Event: *newEvent("initialized")})
					send(&dap.AttachResponse{Response: *newResponse(request.Request)})
					return
				case *dap.DisconnectRequest: // Required
					panic("BAD")
					return
				case *dap.PauseRequest: // Required
					panic("BAD")
					return
				default:
					panic("BAD")
				}
			}
		}()
	}
}

func newEvent(event string) *dap.Event {
	return &dap.Event{
		ProtocolMessage: dap.ProtocolMessage{
			Seq:  0,
			Type: "event",
		},
		Event: event,
	}
}

func newResponse(request dap.Request) *dap.Response {
	return &dap.Response{
		ProtocolMessage: dap.ProtocolMessage{
			Seq:  0,
			Type: "response",
		},
		Command:    request.Command,
		RequestSeq: request.Seq,
		Success:    true,
	}
}
