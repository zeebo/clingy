package clingy

import (
	"strings"

	"github.com/zeebo/errs/v2"
)

type argsHandler struct {
	args    []string
	used    []bool
	dynamic func(string) ([]string, error)
}

func newArgsHandler(args []string, dynamic func(string) ([]string, error)) *argsHandler {
	return &argsHandler{
		args:    args,
		used:    make([]bool, len(args)),
		dynamic: dynamic,
	}
}

func (ah *argsHandler) PeekArgs() []string {
	out := make([]string, 0, len(ah.args))
	for i, arg := range ah.args {
		if ah.used[i] || arg == "--" {
			continue
		}
		out = append(out, arg)
	}
	return out
}

func (ah *argsHandler) ConsumeArgs() ([]string, error) {
	out := ah.PeekArgs()
	for i := range ah.used {
		ah.used[i] = true
	}
	return out, nil
}

func (ah *argsHandler) PeekArg(name string) (string, bool, error) {
	for i, arg := range ah.args {
		if ah.used[i] || arg == "--" {
			continue
		}
		if len(arg) > 1 && arg[0] == '-' {
			return "", false, errs.Tag("argument error").Errorf("unknown flag: %q", arg)
		}
		return arg, true, nil
	}
	return "", false, nil
}

func (ah *argsHandler) ConsumeArg(name string) (string, bool, error) {
	for i, arg := range ah.args {
		if ah.used[i] || arg == "--" {
			continue
		}
		if len(arg) > 1 && arg[0] == '-' {
			return "", false, errs.Tag("argument error").Errorf("unknown flag: %q", arg)
		}
		ah.used[i] = true
		return arg, true, nil
	}
	return "", false, nil
}

func (ah *argsHandler) ConsumeFlag(name string, bstyle bool) (values []string, err error) {
	var used []uint

	for i := uint(0); i < uint(len(ah.args)); i++ {
		arg := ah.args[i]

		// check if the argument ends all flags
		if arg == "--" {
			break
		}

		// check if the argument is positional
		if len(arg) < 1 || arg[0] != '-' || arg == "-" {
			continue
		}

		// strip off the -- prefix for easier processing
		if len(arg) >= 2 && arg[:2] == "--" {
			arg = arg[2:]
		} else {
			arg = arg[1:]
		}

		// check for --foo=bar form
		if idx := strings.IndexByte(arg, '='); idx >= 0 && name == arg[:idx] {
			values = append(values, arg[idx+1:])
			used = append(used, i)
			continue
		}

		// check if the name matches
		if arg != name {
			continue
		}

		// if the flag is boolean style, then the default value is true
		if bstyle {
			values = append(values, "true")
			used = append(used, i)
			continue
		}

		// if we don't have a value specified, we have an error
		if i+1 >= uint(len(ah.args)) || ah.used[i+1] || ah.args[i+1] == "--" {
			return nil, errs.Tag("argument error").Errorf("no value for flag %q", name)
		}

		// consume the next argument as the flag value
		values = append(values, ah.args[i+1])
		used = append(used, i, i+1)
		i++
	}

	// if the flag was not found, try calling the dynamic callback
	if values == nil && ah.dynamic != nil {
		return ah.dynamic(name)
	}

	for _, i := range used {
		ah.used[i] = true
	}

	return values, nil
}
