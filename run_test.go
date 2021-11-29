package clingy_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeebo/clingy"
)

func TestRun_HelpDisplay(t *testing.T) {
	runTestCommand := func(args ...string) (string, error) {
		var stdout bytes.Buffer

		_, err := clingy.Environment{
			Name: "testcommand",
			Args: args,

			Stdout: &stdout,
		}.Run(context.Background(), func(cmds clingy.Commands) {
			cmds.New("subcommand", "test description", nil)
		})
		return stdout.String(), err
	}

	// test help for root command
	result, err := runTestCommand("-h")
	require.NoError(t, err)
	require.Equal(t, `
Usage:
    testcommand [command]

Available commands:
    subcommand    test description

Global flags:
    -h, --help         prints help for the command
        --advanced     when used with -h, prints advanced flags help

Use "testcommand [command] --help" for more information about a command.
`, "\n"+result)

	// test help for subcommand
	result, err = runTestCommand("subcommand", "-h")
	require.NoError(t, err)
	require.Equal(t, `
Usage:
    testcommand subcommand

    test description

Global flags:
    -h, --help         prints help for the command
        --advanced     when used with -h, prints advanced flags help
`, "\n"+result)
}
