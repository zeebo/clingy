package clingy_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/zeebo/clingy"
)

func TestUsage_Exhaustive(t *testing.T) {
	root := &funcCommand{
		SetupFn: func(params clingy.Parameters) {
			n := 0
			desc := func() string { n++; return fmt.Sprint("desc ", n) }

			parseInt := clingy.Transform(strconv.Atoi)
			parseBool := clingy.Transform(strconv.ParseBool)
			parseDuration := clingy.Transform(time.ParseDuration)

			_ = params.Flag("Flags.ValString", desc(), "").(string)
			_ = params.Flag("Flags.ValInt", desc(), 0, parseInt).(int)
			_ = params.Flag("Flags.ValBool", desc(), false, clingy.Boolean, parseBool).(bool)
			_ = params.Flag("Flags.ValDur", desc(), time.Duration(0), parseDuration).(time.Duration)

			params.Break()
			_ = params.Flag("Flags.OptString", desc(), (*string)(nil), clingy.Optional).(*string)
			_ = params.Flag("Flags.OptInt", desc(), (*int)(nil), clingy.Optional, parseInt).(*int)
			_ = params.Flag("Flags.OptBool", desc(), (*bool)(nil), clingy.Optional, clingy.Boolean, parseBool).(*bool)

			params.Break()
			_ = params.Flag("Flags.RepString", desc(), []string(nil), clingy.Repeated).([]string)
			_ = params.Flag("Flags.RepInt", desc(), []int(nil), clingy.Repeated, parseInt).([]int)
			_ = params.Flag("Flags.RepBool", desc(), []bool(nil), clingy.Repeated, clingy.Boolean, parseBool).([]bool)

			params.Break()
			_ = params.Flag("Flags.Def", desc(), "some default").(string)
			_ = params.Flag("Flags.DefRep", desc(), []string{"some default"}, clingy.Repeated).([]string)
			_ = params.Flag("Flags.DefEnv", desc(), "some default", clingy.Getenv("ENV")).(string)
			_ = params.Flag("Flags.DefRepEnv", desc(), []string{"some default"}, clingy.Repeated, clingy.Getenv("ENV")).([]string)

			params.Break()
			_ = params.Flag("Flags.Req", desc(), clingy.Required).(string)
			_ = params.Flag("Flags.ReqRep", desc(), clingy.Required, clingy.Repeated).([]string)
			_ = params.Flag("Flags.ReqEnv", desc(), clingy.Required, clingy.Getenv("ENV")).(string)
			_ = params.Flag("Flags.ReqRepEnv", desc(), clingy.Required, clingy.Repeated, clingy.Getenv("ENV")).([]string)

			params.Break()
			_ = params.Flag("Flags.Custom", desc(), "", clingy.Type("custom")).(string)
			_ = params.Flag("Flags.Short", desc(), "", clingy.Short('s')).(string)
			_ = params.Flag("Flags.Advanced", desc(), "", clingy.Advanced).(string)

			_ = params.Flag("Flags.HiddenString", "", "", clingy.Hidden).(string)
			_ = params.Flag("Flags.HiddenInt", "", 0, parseInt, clingy.Hidden).(int)
			_ = params.Flag("Flags.HiddenBool", "", false, clingy.Boolean, parseBool, clingy.Hidden).(bool)

			_ = params.Arg("Args.ValString", desc()).(string)
			_ = params.Arg("Args.ValInt", desc(), parseInt).(int)

			_ = params.Arg("Args.OptString", desc(), clingy.Optional).(*string)
			_ = params.Arg("Args.OptInt", desc(), clingy.Optional, parseInt).(*int)

			_ = params.Arg("Args.RepInt", desc(), clingy.Repeated, parseInt).([]int)

		},
		ExecuteFn: func(ctx context.Context) error { return nil },
	}

	{
		result := Run(root, "-h")
		result.AssertValid(t)
		result.AssertStdout(t, `
			Usage:
			    testcommand <--Flags.Req string> <--Flags.ReqRep string ...> <--Flags.ReqEnv string> <--Flags.ReqRepEnv string ...> [flags] <Args.ValString> <Args.ValInt> [Args.OptString [Args.OptInt [Args.RepInt ...]]]

			Arguments:
			    Args.ValString    desc 22
			    Args.ValInt       desc 23
			    Args.OptString    desc 24
			    Args.OptInt       desc 25
			    Args.RepInt       desc 26

			Flags:
			        --Flags.ValString string    desc 1
			        --Flags.ValInt int          desc 2
			        --Flags.ValBool             desc 3
			        --Flags.ValDur duration     desc 4

			        --Flags.OptString string    desc 5
			        --Flags.OptInt int          desc 6
			        --Flags.OptBool             desc 7

			        --Flags.RepString string    desc 8 (repeated)
			        --Flags.RepInt int          desc 9 (repeated)
			        --Flags.RepBool             desc 10 (repeated)

			        --Flags.Def string          desc 11 (default "some default")
			        --Flags.DefRep string       desc 12 (repeated) (default [some default])
			        --Flags.DefEnv string       desc 13 (env ENV) (default "some default")
			        --Flags.DefRepEnv string    desc 14 (repeated) (env ENV) (default [some default])

			        --Flags.Req string          desc 15 (required)
			        --Flags.ReqRep string       desc 16 (required) (repeated)
			        --Flags.ReqEnv string       desc 17 (required) (env ENV)
			        --Flags.ReqRepEnv string    desc 18 (required) (repeated) (env ENV)

			        --Flags.Custom custom    desc 19
			    -s, --Flags.Short string     desc 20

			Global flags:
			    -h, --help         prints help for the command
			        --summary      prints a summary of what commands are available
			        --advanced     when used with -h, prints advanced flags help
		`)
	}

	{
		result := Run(root, "-h", "--advanced")
		result.AssertValid(t)
		result.AssertStdout(t, `
			Usage:
			    testcommand [--Flags.ValString string] [--Flags.ValInt int] [--Flags.ValBool] [--Flags.ValDur duration] [--Flags.OptString string] [--Flags.OptInt int] [--Flags.OptBool] [--Flags.RepString string ...] [--Flags.RepInt int ...] [--Flags.RepBool  ...] [--Flags.Def string] [--Flags.DefRep string ...] [--Flags.DefEnv string] [--Flags.DefRepEnv string ...] <--Flags.Req string> <--Flags.ReqRep string ...> <--Flags.ReqEnv string> <--Flags.ReqRepEnv string ...> [--Flags.Custom custom] [--Flags.Short string] [--Flags.Advanced string] <Args.ValString> <Args.ValInt> [Args.OptString [Args.OptInt [Args.RepInt ...]]]

			Arguments:
			    Args.ValString    desc 22
			    Args.ValInt       desc 23
			    Args.OptString    desc 24
			    Args.OptInt       desc 25
			    Args.RepInt       desc 26

			Flags:
			        --Flags.ValString string    desc 1
			        --Flags.ValInt int          desc 2
			        --Flags.ValBool             desc 3
			        --Flags.ValDur duration     desc 4

			        --Flags.OptString string    desc 5
			        --Flags.OptInt int          desc 6
			        --Flags.OptBool             desc 7

			        --Flags.RepString string    desc 8 (repeated)
			        --Flags.RepInt int          desc 9 (repeated)
			        --Flags.RepBool             desc 10 (repeated)

			        --Flags.Def string          desc 11 (default "some default")
			        --Flags.DefRep string       desc 12 (repeated) (default [some default])
			        --Flags.DefEnv string       desc 13 (env ENV) (default "some default")
			        --Flags.DefRepEnv string    desc 14 (repeated) (env ENV) (default [some default])

			        --Flags.Req string          desc 15 (required)
			        --Flags.ReqRep string       desc 16 (required) (repeated)
			        --Flags.ReqEnv string       desc 17 (required) (env ENV)
			        --Flags.ReqRepEnv string    desc 18 (required) (repeated) (env ENV)

			        --Flags.Custom custom      desc 19
			    -s, --Flags.Short string       desc 20
			        --Flags.Advanced string    desc 21

			Global flags:
			    -h, --help         prints help for the command
			        --summary      prints a summary of what commands are available
			        --advanced     when used with -h, prints advanced flags help
		`)
	}
}

func TestUsage_MultilineDescription(t *testing.T) {
	result := Capture(Env("testcommand", nil, "command", "-h"), func(cmds clingy.Commands) {
		cmds.New("command", `single line description
			The longer multiline description is here.
			It contains multiple lines. Neato.
		`, nil)
	})
	result.AssertValid(t)
	result.AssertStdout(t, `
		Usage:
		    testcommand command

		    single line description

		    The longer multiline description is here.
		    It contains multiple lines. Neato.

		Global flags:
		    -h, --help         prints help for the command
		        --summary      prints a summary of what commands are available
		        --advanced     when used with -h, prints advanced flags help
	`)
}

func TestUsage_DistanceSuggestions(t *testing.T) {
	cmds := func(cmds clingy.Commands) {
		cmds.New("cmd1", "d1", nil)
		cmds.New("cmd2", "d2", nil)
		cmds.New("cmb3", "d3", nil)
		cmds.Group("grp1", "g1", func() {
			cmds.New("cmd4", "d4", nil)
		})
	}

	{
		env := Env("testcommand", nil, "amd4")

		result := Capture(env, cmds)
		result.AssertStdout(t, `
			Errors:
			    unknown command: "amd4". did you mean:
			        cmd1
			        cmd2

			Usage:
			    testcommand [command]

			Available commands:
			    cmd1    d1
			    cmd2    d2
			    cmb3    d3
			    grp1    g1

			Global flags:
			    -h, --help         prints help for the command
			        --summary      prints a summary of what commands are available
			        --advanced     when used with -h, prints advanced flags help

			Use "testcommand [command] --help" for more information about a command.
		`)
	}

	{
		env := Env("testcommand", nil, "grp1", "--foo")

		result := Capture(env, cmds)
		result.AssertStdout(t, `
			Errors:
			    argument error: unknown flag: "--foo"

			Usage:
			    testcommand grp1 [command]

			    g1

			Available commands:
			    cmd4    d4

			Global flags:
			    -h, --help         prints help for the command
			        --summary      prints a summary of what commands are available
			        --advanced     when used with -h, prints advanced flags help

			Use "testcommand grp1 [command] --help" for more information about a command.
		`)
	}

	{
		env := Env("testcommand", nil, "amd4")
		env.SuggestionsMinEditDistance = -1

		result := Capture(env, cmds)
		result.AssertStdout(t, `
			Errors:
			    unknown command: "amd4"

			Usage:
			    testcommand [command]

			Available commands:
			    cmd1    d1
			    cmd2    d2
			    cmb3    d3
			    grp1    g1

			Global flags:
			    -h, --help         prints help for the command
			        --summary      prints a summary of what commands are available
			        --advanced     when used with -h, prints advanced flags help

			Use "testcommand [command] --help" for more information about a command.
		`)
	}

	{
		env := Env("testcommand", nil, "grp1", "--foo")
		env.SuggestionsMinEditDistance = -1

		result := Capture(env, cmds)
		result.AssertStdout(t, `
			Errors:
			    argument error: unknown flag: "--foo"

			Usage:
			    testcommand grp1 [command]

			    g1

			Available commands:
			    cmd4    d4

			Global flags:
			    -h, --help         prints help for the command
			        --summary      prints a summary of what commands are available
			        --advanced     when used with -h, prints advanced flags help

			Use "testcommand grp1 [command] --help" for more information about a command.
		`)
	}
}
