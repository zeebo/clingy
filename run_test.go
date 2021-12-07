package clingy_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/zeebo/assert"
	"github.com/zeebo/clingy"
)

type command struct {
}

func (cmd *command) Setup(params clingy.Parameters) {
	params.Arg("paramA", "paramA description")
}

func (cmd *command) Execute(ctx clingy.Context) error {
	return nil
}

func TestRun_HelpDisplay(t *testing.T) {
	runTestCommand := func(args ...string) (string, error) {
		var stdout bytes.Buffer

		_, err := clingy.Environment{
			Name: "testcommand",
			Args: args,

			Stdout: &stdout,
		}.Run(context.Background(), func(cmds clingy.Commands) {
			cmds.New("subcommand", "test description", &command{})
		})
		return stdout.String(), err
	}

	// test help for root command
	result, err := runTestCommand("-h")
	assert.NoError(t, err)
	assert.Equal(t, `
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
	assert.NoError(t, err)
	assert.Equal(t, `
Usage:
    testcommand subcommand <paramA>

    test description

Arguments:
    paramA    paramA description

Global flags:
    -h, --help         prints help for the command
        --advanced     when used with -h, prints advanced flags help
`, "\n"+result)

	// test help for subcommand without mandatory parameter
	result, err = runTestCommand("subcommand")
	assert.NoError(t, err)
	assert.Equal(t, `
Errors:
    argument error: paramA: required argument missing

Usage:
    testcommand subcommand <paramA>

    test description

Arguments:
    paramA    paramA description

Global flags:
    -h, --help         prints help for the command
        --advanced     when used with -h, prints advanced flags help
`, "\n"+result)
}
