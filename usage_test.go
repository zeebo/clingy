package clingy_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/zeebo/clingy"
)

func TestRun_UsageExhaustive(t *testing.T) {
	root := &funcCommand{
		SetupFn: func(params clingy.Parameters) {
			n := 0
			desc := func() string { n++; return fmt.Sprint("desc ", n) }

			parseInt := clingy.Transform(strconv.Atoi)
			parseBool := clingy.Transform(strconv.ParseBool)

			_ = params.Flag("Flags.ValString", desc(), "").(string)
			_ = params.Flag("Flags.ValInt", desc(), 0, parseInt).(int)
			_ = params.Flag("Flags.ValBool", desc(), false, clingy.Boolean, parseBool).(bool)

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
			    Args.ValString    desc 21
			    Args.ValInt       desc 22
			    Args.OptString    desc 23
			    Args.OptInt       desc 24
			    Args.RepInt       desc 25

			Flags:
			        --Flags.ValString string    desc 1
			        --Flags.ValInt int          desc 2
			        --Flags.ValBool             desc 3

			        --Flags.OptString string    desc 4
			        --Flags.OptInt int          desc 5
			        --Flags.OptBool             desc 6

			        --Flags.RepString string    desc 7 (repeated)
			        --Flags.RepInt int          desc 8 (repeated)
			        --Flags.RepBool             desc 9 (repeated)

			        --Flags.Def string          desc 10 (default "some default")
			        --Flags.DefRep string       desc 11 (repeated) (default [some default])
			        --Flags.DefEnv string       desc 12 (env ENV) (default "some default")
			        --Flags.DefRepEnv string    desc 13 (repeated) (env ENV) (default [some default])

			        --Flags.Req string          desc 14 (required)
			        --Flags.ReqRep string       desc 15 (required) (repeated)
			        --Flags.ReqEnv string       desc 16 (required) (env ENV)
			        --Flags.ReqRepEnv string    desc 17 (required) (repeated) (env ENV)

			        --Flags.Custom custom    desc 18
			    -s, --Flags.Short string     desc 19

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
			    testcommand [--Flags.ValString string] [--Flags.ValInt int] [--Flags.ValBool] [--Flags.OptString string] [--Flags.OptInt int] [--Flags.OptBool] [--Flags.RepString string ...] [--Flags.RepInt int ...] [--Flags.RepBool  ...] [--Flags.Def string] [--Flags.DefRep string ...] [--Flags.DefEnv string] [--Flags.DefRepEnv string ...] <--Flags.Req string> <--Flags.ReqRep string ...> <--Flags.ReqEnv string> <--Flags.ReqRepEnv string ...> [--Flags.Custom custom] [--Flags.Short string] [--Flags.Advanced string] <Args.ValString> <Args.ValInt> [Args.OptString [Args.OptInt [Args.RepInt ...]]]

			Arguments:
			    Args.ValString    desc 21
			    Args.ValInt       desc 22
			    Args.OptString    desc 23
			    Args.OptInt       desc 24
			    Args.RepInt       desc 25

			Flags:
			        --Flags.ValString string    desc 1
			        --Flags.ValInt int          desc 2
			        --Flags.ValBool             desc 3

			        --Flags.OptString string    desc 4
			        --Flags.OptInt int          desc 5
			        --Flags.OptBool             desc 6

			        --Flags.RepString string    desc 7 (repeated)
			        --Flags.RepInt int          desc 8 (repeated)
			        --Flags.RepBool             desc 9 (repeated)

			        --Flags.Def string          desc 10 (default "some default")
			        --Flags.DefRep string       desc 11 (repeated) (default [some default])
			        --Flags.DefEnv string       desc 12 (env ENV) (default "some default")
			        --Flags.DefRepEnv string    desc 13 (repeated) (env ENV) (default [some default])

			        --Flags.Req string          desc 14 (required)
			        --Flags.ReqRep string       desc 15 (required) (repeated)
			        --Flags.ReqEnv string       desc 16 (required) (env ENV)
			        --Flags.ReqRepEnv string    desc 17 (required) (repeated) (env ENV)

			        --Flags.Custom custom      desc 18
			    -s, --Flags.Short string       desc 19
			        --Flags.Advanced string    desc 20

			Global flags:
			    -h, --help         prints help for the command
			        --summary      prints a summary of what commands are available
			        --advanced     when used with -h, prints advanced flags help
		`)
	}
}
