package clingy_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/zeebo/assert"
	"github.com/zeebo/clingy"
)

type command struct{}

func (cmd *command) Setup(params clingy.Parameters) {
	params.Arg("paramA", "paramA description")
}

func (cmd *command) Execute(ctx context.Context) error {
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

func TestRun(t *testing.T) {
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

	t.Run("BasicCalls", func(t *testing.T) {
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
	})

	t.Run("MissingCalls", func(t *testing.T) {
		for _, cmd := range [][]string{
			{"cmd5"},
			{"group1", "sub2"},
			{"group1", "group2", "sub3"},
		} {
			result := clingy.Capture(clingy.Env("cmd", cmd...), cmds)
			result.AssertStdoutContains(t, "unknown command")
			assert.That(t, !result.Ok)
		}
	})
}

type stdioCommand struct {
	ExtraOutput string
}

func (cmd *stdioCommand) Setup(params clingy.Parameters) {}

func (cmd *stdioCommand) Execute(ctx context.Context) error {
	in, _ := io.ReadAll(clingy.Stdin(ctx))
	clingy.Stdout(ctx).Write(in)
	clingy.Stdout(ctx).Write([]byte(cmd.ExtraOutput))
	clingy.Stderr(ctx).Write(in)
	clingy.Stderr(ctx).Write([]byte(cmd.ExtraOutput))
	return nil
}

func TestRun_Stdio(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	_, err := clingy.Environment{
		Name: "testcommand",
		Args: []string{"run"},

		Stdin:  strings.NewReader("hello world"),
		Stdout: &stdout,
		Stderr: &stderr,
	}.Run(context.Background(), func(cmds clingy.Commands) {
		cmds.New("run", "check stdio", &stdioCommand{})
	})

	assert.NoError(t, err)
	assert.Equal(t, stdout.String(), "hello world")
	assert.Equal(t, stderr.String(), "hello world")
}

func TestRun_Root(t *testing.T) {
	cmd1 := &stdioCommand{ExtraOutput: "cmd1"}
	cmd2 := &stdioCommand{ExtraOutput: "cmd2"}
	root := &stdioCommand{ExtraOutput: "root"}

	check := func(expected string, args ...string) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		_, err := clingy.Environment{
			Root: root,
			Name: "testcommand",
			Args: append([]string{}, args...),

			Stdin:  strings.NewReader(""),
			Stdout: &stdout,
			Stderr: &stderr,
		}.Run(context.Background(), func(cmds clingy.Commands) {
			cmds.New("cmd1", "cmd1", cmd1)
			cmds.New("cmd2", "cmd2", cmd2)
		})

		assert.NoError(t, err)
		assert.Equal(t, stdout.String(), expected)
		assert.Equal(t, stderr.String(), expected)
	}

	check("root")
	check("cmd1", "cmd1")
	check("cmd2", "cmd2")
}

func TestRun_RootHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	ok, err := clingy.Environment{
		Root: clingy.NewRecordingCmd("root"),
		Name: "testcommand",
		Args: []string{"-h"},

		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	}.Run(context.Background(), nil)

	assert.NoError(t, err)
	assert.That(t, ok)
	assert.Equal(t, "", stderr.String())

	assert.Equal(t, `
Usage:
    testcommand [flags] <Args.ValString> <Args.ValInt> [Args.OptString [Args.OptInt [Args.RepInt ...]]]

Arguments:
    Args.ValString    
    Args.ValInt       
    Args.OptString    
    Args.OptInt       
    Args.RepInt       

Flags:
        --Flags.ValString string    
        --Flags.ValInt int          
        --Flags.ValBool             
        --Flags.OptString string    
        --Flags.OptInt int          
        --Flags.OptBool             
        --Flags.RepString string     (repeated)
        --Flags.RepInt int           (repeated)
        --Flags.RepBool              (repeated)

Global flags:
    -h, --help         prints help for the command
        --advanced     when used with -h, prints advanced flags help
`, "\n"+stdout.String())
}

func TestRun_HiddenParse(t *testing.T) {
	root := clingy.NewRecordingCmd("root")

	ok, err := clingy.Environment{
		Root: root,
		Name: "testcommand",
		Args: []string{
			"arg1", "10",
			"--Flags.HiddenInt", "5",
			"--Flags.HiddenString", "foo",
			"--Flags.HiddenBool",
		},

		Stdin: strings.NewReader(""),
	}.Run(context.Background(), nil)

	assert.NoError(t, err)
	assert.That(t, ok)

	assert.Equal(t, root.Flags.HiddenInt, 5)
	assert.Equal(t, root.Flags.HiddenString, "foo")
	assert.Equal(t, root.Flags.HiddenBool, true)
}

type setupFailCommand struct{}

func (cmd *setupFailCommand) Setup(params clingy.Parameters) {
	params.Arg("argument", "failing argument", clingy.Transform(func(_ string) (string, error) {
		return "", errors.New("parse failure")
	}))
}

func (cmd *setupFailCommand) Execute(ctx context.Context) error {
	return errors.New("unreachable")
}

func TestRun_InputValidation(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	ok, err := clingy.Environment{
		Root: new(setupFailCommand),
		Name: "setup-fail",
		Args: []string{"foo"},

		Stdout: &stdout,
		Stderr: &stderr,
	}.Run(context.Background(), nil)

	assert.That(t, !ok)
	assert.That(t, err == nil)
}

type funcCommand struct {
	SetupFn   func(params clingy.Parameters)
	ExecuteFn func(ctx context.Context) error
}

func (cmd *funcCommand) Setup(params clingy.Parameters)    { cmd.SetupFn(params) }
func (cmd *funcCommand) Execute(ctx context.Context) error { return cmd.ExecuteFn(ctx) }

func TestRun_OptionalPtrDeref(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	ok, err := clingy.Environment{
		Root: &funcCommand{
			SetupFn: func(params clingy.Parameters) {
				params.Flag("test", "test flag", new(bool),
					clingy.Transform(strconv.ParseBool), clingy.Boolean, clingy.Optional)
			},
			ExecuteFn: func(ctx context.Context) error { return nil },
		},
		Name: "testcommand",
		Args: []string{"-h"},

		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	}.Run(context.Background(), nil)

	assert.NoError(t, err)
	assert.That(t, ok)
	assert.Equal(t, "", stderr.String())

	assert.Equal(t, `
Usage:
    testcommand [flags]

Flags:
        --test     test flag (default false)

Global flags:
    -h, --help         prints help for the command
        --advanced     when used with -h, prints advanced flags help
`, "\n"+stdout.String())

}
