package clingy_test

import (
	"bytes"
	"context"
	"strings"
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

func TestRun_BasicCalls(t *testing.T) {
	cmds := func(cmds *clingy.RecordingCmds) {
		cmds.New("cmd1")
		cmds.New("cmd2")
		cmds.Group("group1", func() {
			cmds.New("sub1")
			cmds.Group("group2", func() {
				cmds.New("sub2")
			})
			cmds.New("sub3")
		})
		cmds.New("cmd3")
		cmds.New("cmd4")
	}

	for _, cmd := range [][]string{
		{"cmd1"},
		{"cmd2"},
		{"group1", "sub1"},
		{"group1", "group2", "sub2"},
		{"group1", "sub3"},
		{"cmd3"},
		{"cmd4"},
	} {
		name := strings.Join(cmd, " ")
		cmd = append(cmd, "argString", "10")

		result := clingy.Capture(clingy.Env("cmd", cmd...), cmds)
		result.AssertRunValid(t)
		result.AssertExecuted(t, name)
	}
}
