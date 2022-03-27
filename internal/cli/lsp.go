package cli

// LspCommand is the command to test the Solidity project
type LspCommand struct {
	*baseCommand
}

// Help implements the cli.Command interface
func (b *LspCommand) Help() string {
	return `Usage: greenhouse test

  Start the lsp server`
}

// Synopsis implements the cli.Command interface
func (b *LspCommand) Synopsis() string {
	return "Start the lsp server"
}

// Run implements the cli.Command interface
func (b *LspCommand) Run(args []string) int {
	return 0
}
