package clingy_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/zeebo/assert"
	"github.com/zeebo/clingy"
	"github.com/zeebo/errs/v2"
)

func TestRun_HelpDisplay(t *testing.T) {
	run := func(args ...string) Result {
		return Capture(Env("testcommand", nil, args...), func(cmds clingy.Commands) {
			cmds.New("subcommand", "test description", &funcCommand{
				SetupFn:   func(params clingy.Parameters) { params.Arg("paramA", "paramA description") },
				ExecuteFn: func(ctx context.Context) error { return nil },
			})
		})
	}

	{ // test help for root command
		result := run("-h")
		result.AssertValid(t)
		result.AssertStdout(t, `
			Usage:
			    testcommand [command]

			Available commands:
			    subcommand    test description

			Global flags:
			    -h, --help         prints help for the command
			        --summary      prints a summary of what commands are available
			        --advanced     when used with -h, prints advanced flags help

			Use "testcommand [command] --help" for more information about a command.
		`)
	}

	{ // test help for subcommand
		result := run("subcommand", "-h")
		result.AssertValid(t)
		result.AssertStdout(t, `
			Usage:
			    testcommand subcommand <paramA>

			    test description

			Arguments:
			    paramA    paramA description

			Global flags:
			    -h, --help         prints help for the command
			        --summary      prints a summary of what commands are available
			        --advanced     when used with -h, prints advanced flags help
		`)
	}

	{ // test help for subcommand without mandatory parameter
		result := run("subcommand")
		assert.That(t, !result.Ok)
		assert.That(t, result.Err == nil)
		result.AssertStdout(t, `
			Errors:
			    argument error: paramA: required argument missing

			Usage:
			    testcommand subcommand <paramA>

			    test description

			Arguments:
			    paramA    paramA description

			Global flags:
			    -h, --help         prints help for the command
			        --summary      prints a summary of what commands are available
			        --advanced     when used with -h, prints advanced flags help
		`)
	}
}

func TestRun_Dispatches(t *testing.T) {
	cmds := func(cmds clingy.Commands) {
		cmds.New("cmd1", "", printCommand("cmd1"))
		cmds.New("cmd2", "", printCommand("cmd2"))
		cmds.Group("group1", "", func() {
			cmds.New("sub1", "", printCommand("group1 sub1"))
			cmds.Group("group2", "", func() {
				cmds.New("sub2", "", printCommand("group1 group2 sub2"))
			})
			cmds.New("sub3", "", printCommand("group1 sub3"))
		})
		cmds.New("cmd3", "", printCommand("cmd3"))
		cmds.New("cmd4", "", printCommand("cmd4"))
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
			result := Capture(Env("cmd", nil, cmd...), cmds)
			result.AssertValid(t)
			result.AssertStdout(t, strings.Join(cmd, " "))
		}
	})

	t.Run("MissingCalls", func(t *testing.T) {
		for _, cmd := range [][]string{
			{"cmd5"},
			{"group1", "sub2"},
			{"group1", "group2", "sub3"},
		} {
			result := Capture(Env("cmd", nil, cmd...), cmds)
			assert.That(t, !result.Ok)
			result.AssertStdoutContains(t, `unknown command: `)
		}
	})
}

func TestRun_Stdio(t *testing.T) {
	env := Env("testcommand", nil, "run")
	env.Stdin = strings.NewReader("hello world")

	result := Capture(env, func(cmds clingy.Commands) {
		cmds.New("run", "check stdio", &funcCommand{
			SetupFn: func(params clingy.Parameters) {},
			ExecuteFn: func(ctx context.Context) error {
				in, _ := io.ReadAll(clingy.Stdin(ctx))
				clingy.Stdout(ctx).Write(in)
				clingy.Stderr(ctx).Write(in)
				return nil
			},
		})
	})
	result.AssertValid(t)
	result.AssertStdout(t, "hello world")
	result.AssertStderr(t, "hello world")
}

func TestRun_Root(t *testing.T) {
	check := func(expected string, args ...string) {
		result := Capture(Env("testcommand", printCommand("root"), args...),
			func(cmds clingy.Commands) {
				cmds.New("cmd1", "cmd1", printCommand("cmd1"))
				cmds.New("cmd2", "cmd2", printCommand("cmd2"))
			})
		result.AssertValid(t)
		result.AssertStdout(t, expected)
	}

	check("root")
	check("cmd1", "cmd1")
	check("cmd2", "cmd2")
}

func TestRun_HiddenParse(t *testing.T) {
	var (
		fstring string
		fint    int
		fbool   bool
	)

	root := &funcCommand{
		SetupFn: func(params clingy.Parameters) {
			parseInt := clingy.Transform(strconv.Atoi)
			parseBool := clingy.Transform(strconv.ParseBool)

			fstring = params.Flag("Flags.HiddenString", "", "", clingy.Hidden).(string)
			fint = params.Flag("Flags.HiddenInt", "", 0, parseInt, clingy.Hidden).(int)
			fbool = params.Flag("Flags.HiddenBool", "", false, clingy.Boolean, parseBool, clingy.Hidden).(bool)
		},
		ExecuteFn: func(ctx context.Context) error { return nil },
	}

	result := Run(root,
		"--Flags.HiddenInt", "5",
		"--Flags.HiddenString", "foo",
		"--Flags.HiddenBool",
	)
	result.AssertValid(t)

	assert.Equal(t, fint, 5)
	assert.Equal(t, fstring, "foo")
	assert.Equal(t, fbool, true)
}

func TestRun_InputValidation(t *testing.T) {
	root := &funcCommand{
		ExecuteFn: func(ctx context.Context) error { return errors.New("unreachable") },
		SetupFn: func(params clingy.Parameters) {
			params.Arg("argument", "failing argument", clingy.Transform(func(_ string) (string, error) {
				return "", errors.New("parse failure")
			}))
		},
	}

	result := Run(root, "foo")
	assert.That(t, !result.Ok)
	assert.That(t, result.Err == nil)
}

func TestRun_OptionalPtrDeref(t *testing.T) {
	root := &funcCommand{
		ExecuteFn: func(ctx context.Context) error { return nil },
		SetupFn: func(params clingy.Parameters) {
			params.Flag("test", "test flag", new(bool), clingy.Transform(strconv.ParseBool), clingy.Boolean, clingy.Optional)
		},
	}

	result := Run(root, "-h")
	result.AssertValid(t)
	result.AssertStdout(t, `
		Usage:
		    testcommand [flags]

		Flags:
		        --test     test flag (default false)

		Global flags:
		    -h, --help         prints help for the command
		        --summary      prints a summary of what commands are available
		        --advanced     when used with -h, prints advanced flags help
	`)

}

func TestRun_GetenvUsage(t *testing.T) {
	root := &funcCommand{
		SetupFn:   func(params clingy.Parameters) { params.Flag("test", "test flag", "", clingy.Getenv("TEST_FLAG")) },
		ExecuteFn: func(ctx context.Context) error { return nil },
	}

	result := Run(root, "-h")
	result.AssertValid(t)
	result.AssertStdout(t, `
		Usage:
		    testcommand [flags]

		Flags:
		        --test string    test flag (env TEST_FLAG)

		Global flags:
		    -h, --help         prints help for the command
		        --summary      prints a summary of what commands are available
		        --advanced     when used with -h, prints advanced flags help
	`)
}

func TestRun_Getenv(t *testing.T) {
	var flagval string

	root := &funcCommand{
		SetupFn: func(params clingy.Parameters) {
			flagval = params.Flag("test", "test flag", "", clingy.Getenv("TEST_FLAG")).(string)
		},
		ExecuteFn: func(ctx context.Context) error {
			fmt.Fprint(clingy.Stdout(ctx), flagval)
			return nil
		},
	}

	env := Env("testcommand", root)
	env.Getenv = func(key string) string { return "got from env" }

	result := Capture(env, nil)
	result.AssertValid(t)
	result.AssertStdout(t, "got from env")
}

func TestRun_RequiredFlag(t *testing.T) {
	root := &funcCommand{
		SetupFn:   func(params clingy.Parameters) { params.Flag("test", "test flag", clingy.Required) },
		ExecuteFn: func(ctx context.Context) error { return nil },
	}

	result := Run(root)
	assert.That(t, !result.Ok)
	assert.That(t, result.Err == nil)
	result.AssertStdout(t, `
		Errors:
		    argument error: test: required flag missing

		Usage:
		    testcommand <--test string>

		Flags:
		        --test string    test flag (required)

		Global flags:
		    -h, --help         prints help for the command
		        --summary      prints a summary of what commands are available
		        --advanced     when used with -h, prints advanced flags help
	`)
}

func TestRun_ExtraArguments(t *testing.T) {
	root := &funcCommand{
		SetupFn:   func(params clingy.Parameters) { params.Arg("arg", "some argument") },
		ExecuteFn: func(ctx context.Context) error { return nil },
	}

	result := Run(root, "foo", "bar")
	assert.That(t, !result.Ok)
	assert.That(t, result.Err == nil)
	result.AssertStdout(t, `
		Errors:
		    argument error: unknown arguments: ["bar"]

		Usage:
		    testcommand <arg>

		Arguments:
		    arg    some argument

		Global flags:
		    -h, --help         prints help for the command
		        --summary      prints a summary of what commands are available
		        --advanced     when used with -h, prints advanced flags help
	`)
}

func TestRun_ExtraFlag(t *testing.T) {
	{ // flag coming after repeated arguments
		root := &funcCommand{
			SetupFn:   func(params clingy.Parameters) { params.Arg("arg", "some argument", clingy.Repeated) },
			ExecuteFn: func(ctx context.Context) error { return nil },
		}

		result := Run(root, "foo", "bar", "--baz")
		assert.That(t, !result.Ok)
		assert.That(t, result.Err == nil)
		result.AssertStdout(t, `
			Errors:
			    argument error: unknown flag: "--baz"

			Usage:
			    testcommand [arg ...]

			Arguments:
			    arg    some argument

			Global flags:
			    -h, --help         prints help for the command
			        --summary      prints a summary of what commands are available
			        --advanced     when used with -h, prints advanced flags help
		`)
	}

	{ // flag coming after single argument
		root := &funcCommand{
			SetupFn:   func(params clingy.Parameters) { params.Arg("arg", "some argument") },
			ExecuteFn: func(ctx context.Context) error { return nil },
		}

		result := Run(root, "val", "--baz")
		assert.That(t, !result.Ok)
		assert.That(t, result.Err == nil)
		result.AssertStdout(t, `
			Errors:
			    argument error: unknown flag: "--baz"

			Usage:
			    testcommand <arg>

			Arguments:
			    arg    some argument

			Global flags:
			    -h, --help         prints help for the command
			        --summary      prints a summary of what commands are available
			        --advanced     when used with -h, prints advanced flags help
		`)
	}
}

func TestRun_InvalidFlags(t *testing.T) {
	root := &funcCommand{
		SetupFn: func(params clingy.Parameters) {
			_ = params.Flag("flag", "some int flag", nil, clingy.Transform(strconv.Atoi), clingy.Short('f')).(int)
		},
		ExecuteFn: func(ctx context.Context) error { return nil },
	}

	{ // nil default grabs zero value from transform
		result := Run(root)
		result.AssertValid(t)
		result.AssertStdout(t, ``)
	}

	{ // no value passed
		result := Run(root, "--flag")
		assert.That(t, !result.Ok)
		result.AssertStdout(t, `
			Errors:
			    argument error: no value for flag "flag"

			Usage:
			    testcommand [flags]

			Flags:
			    -f, --flag int    some int flag

			Global flags:
			    -h, --help         prints help for the command
			        --summary      prints a summary of what commands are available
			        --advanced     when used with -h, prints advanced flags help
		`)
	}

	{ // no value passed with short
		result := Run(root, "-f")
		assert.That(t, !result.Ok)
		result.AssertStdout(t, `
			Errors:
			    argument error: no value for flag "f"

			Usage:
			    testcommand [flags]

			Flags:
			    -f, --flag int    some int flag

			Global flags:
			    -h, --help         prints help for the command
			        --summary      prints a summary of what commands are available
			        --advanced     when used with -h, prints advanced flags help
		`)
	}

	{ // invalid value passed
		result := Run(root, "--flag", "foo")
		assert.That(t, !result.Ok)
		result.AssertStdout(t, `
			Errors:
			    argument error: strconv.Atoi: parsing "foo": invalid syntax

			Usage:
			    testcommand [flags]

			Flags:
			    -f, --flag int    some int flag

			Global flags:
			    -h, --help         prints help for the command
			        --summary      prints a summary of what commands are available
			        --advanced     when used with -h, prints advanced flags help
		`)
	}

	{ // invalid value passed with short
		result := Run(root, "-f", "foo")
		assert.That(t, !result.Ok)
		result.AssertStdout(t, `
			Errors:
			    argument error: strconv.Atoi: parsing "foo": invalid syntax

			Usage:
			    testcommand [flags]

			Flags:
			    -f, --flag int    some int flag

			Global flags:
			    -h, --help         prints help for the command
			        --summary      prints a summary of what commands are available
			        --advanced     when used with -h, prints advanced flags help
		`)
	}
}

func TestRun_InvalidArg(t *testing.T) {
	root := &funcCommand{
		SetupFn: func(params clingy.Parameters) {
			_ = params.Arg("arg", "some int arg", clingy.Transform(strconv.Atoi)).(int)
		},
		ExecuteFn: func(ctx context.Context) error { return nil },
	}

	{ // flag value passed
		result := Run(root, "--flag")
		assert.That(t, !result.Ok)
		result.AssertStdout(t, `
			Errors:
			    argument error: unknown flag: "--flag"

			Usage:
			    testcommand <arg>

			Arguments:
			    arg    some int arg

			Global flags:
			    -h, --help         prints help for the command
			        --summary      prints a summary of what commands are available
			        --advanced     when used with -h, prints advanced flags help
		`)
	}

	{ // invalid value passed
		result := Run(root, "foo")
		assert.That(t, !result.Ok)
		result.AssertStdout(t, `
			Errors:
			    argument error: strconv.Atoi: parsing "foo": invalid syntax

			Usage:
			    testcommand <arg>

			Arguments:
			    arg    some int arg

			Global flags:
			    -h, --help         prints help for the command
			        --summary      prints a summary of what commands are available
			        --advanced     when used with -h, prints advanced flags help
		`)
	}

	{ // repeated value, one valid one invalid
		root := &funcCommand{
			SetupFn: func(params clingy.Parameters) {
				_ = params.Arg("arg", "some int arg", clingy.Transform(strconv.Atoi), clingy.Repeated).([]int)
			},
			ExecuteFn: func(ctx context.Context) error { return nil },
		}

		result := Run(root, "8", "foo")
		assert.That(t, !result.Ok)
		result.AssertStdout(t, `
			Errors:
			    argument error: strconv.Atoi: parsing "foo": invalid syntax

			Usage:
			    testcommand [arg ...]

			Arguments:
			    arg    some int arg

			Global flags:
			    -h, --help         prints help for the command
			        --summary      prints a summary of what commands are available
			        --advanced     when used with -h, prints advanced flags help
		`)
	}

}

func TestRun_SetupPanics(t *testing.T) {
	panics := func(cb func(params clingy.Parameters)) (out string) {
		defer func() { out, _ = recover().(string) }()
		_ = Run(&funcCommand{SetupFn: cb})
		return ""
	}

	{
		result := panics(func(params clingy.Parameters) {
			params.Arg("arg", "some argument")
			params.Flag("flag", "some flag", nil)
		})
		assert.Equal(t, result, "must perform all Flag/Break calls before any Arg calls")
	}

	{
		result := panics(func(params clingy.Parameters) {
			params.Arg("arg", "some argument")
			params.Break()
		})
		assert.Equal(t, result, "must perform all Flag/Break calls before any Arg calls")
	}

	{
		result := panics(func(params clingy.Parameters) {
			params.Arg("foo", "some argument")
			params.Arg("foo", "some argument again")
		})
		assert.Equal(t, result, `parameter already defined with name: "foo"`)
	}

	{
		result := panics(func(params clingy.Parameters) {
			params.Flag("foo", "some flag", 0, clingy.Short('f'))
			params.Flag("bar", "some flag again", 0, clingy.Short('f'))
		})
		assert.Equal(t, result, `parameter already defined with short-name: 'f'`)
	}

	{
		result := panics(func(params clingy.Parameters) {
			params.Flag("foo", "some flag", 0, clingy.Transform(func() bool { return false }))
		})
		assert.Equal(t, result, `parameter has invalid transformation functions: transform: func() bool cannot be applied to string`)
	}

	{
		result := panics(func(params clingy.Parameters) {
			params.Flag("foo", "some flag", clingy.Required, clingy.Optional)
		})
		assert.Equal(t, result, `optional flag with Required default value: "foo"`)
	}

	{
		result := panics(func(params clingy.Parameters) {
			params.Arg("foo", "some flag", clingy.Optional)
			params.Arg("bar", "some flag")
		})
		assert.Equal(t, result, `required argument after optional arguments: "bar"`)
	}

	{
		result := panics(func(params clingy.Parameters) {
			params.Arg("foo", "some flag", clingy.Repeated)
			params.Arg("bar", "some flag")
		})
		assert.Equal(t, result, `argument after repeated argument: "bar"`)
	}
}

func TestRun_Wrap(t *testing.T) {
	root := &funcCommand{
		SetupFn:   func(params clingy.Parameters) {},
		ExecuteFn: func(ctx context.Context) error { return nil },
	}

	env := Env("testcommand", root)
	env.Wrap = func(ctx context.Context, cmd clingy.Command) (err error) {
		assert.Equal(t, cmd, root)
		return errs.Tag("sentinel")
	}

	result := Capture(env, nil)
	assert.That(t, errors.Is(result.Err, errs.Tag("sentinel")))
}
