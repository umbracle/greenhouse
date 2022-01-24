package cli

// TestCommand is the command to test the Solidity project
type TestCommand struct {
	*baseCommand
}

// Help implements the cli.Command interface
func (b *TestCommand) Help() string {
	return `Usage: greenhouse test

  Test the project`
}

// Synopsis implements the cli.Command interface
func (b *TestCommand) Synopsis() string {
	return "Test the project"
}

// Run implements the cli.Command interface
func (b *TestCommand) Run(args []string) int {
	flags := b.Flags("test")
	if err := flags.Parse(args); err != nil {
		b.UI.Error(err.Error())
		return 1
	}

	if err := b.Init(); err != nil {
		b.UI.Error(err.Error())
		return 1
	}
	if err := b.project.Test(); err != nil {
		b.UI.Error(err.Error())
		return 1
	}
	return 0
}
