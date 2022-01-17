package cli

// BuildCommand is the command to show the version of the agent
type BuildCommand struct {
	*baseCommand
}

// Help implements the cli.Command interface
func (b *BuildCommand) Help() string {
	return `Usage: greenhouse build

  Build and compile the project`
}

// Synopsis implements the cli.Command interface
func (b *BuildCommand) Synopsis() string {
	return "Build and compile the project"
}

// Run implements the cli.Command interface
func (b *BuildCommand) Run(args []string) int {
	flags := b.Flags("build")
	if err := flags.Parse(args); err != nil {
		b.UI.Error(err.Error())
		return 1
	}

	if err := b.Init(); err != nil {
		b.UI.Error(err.Error())
		return 1
	}
	if err := b.project.Compile(); err != nil {
		b.UI.Error(err.Error())
		return 1
	}

	b.UI.Output("Compiled.")
	return 0
}
