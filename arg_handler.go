package clingy

import (
	"strings"

	"github.com/zeebo/errs/v2"
)

type argsHandler struct {
	args    []string
	used    []bool
	dynamic func(string) ([]string, error)
	getenv  func(string) string
}

func newArgsHandler(args []string, dynamic func(string) ([]string, error), getenv func(string) string) *argsHandler {
	return &argsHandler{
		args:    args,
		used:    make([]bool, len(args)),
		dynamic: dynamic,
		getenv:  getenv,
	}
}

// walkArgs visits unused positional arguments until visit returns false.
// Unknown flags are either skipped (for PeekArgs) or reported as errors.
func (ah *argsHandler) walkArgs(skipFlags bool, visit func(int) bool) error {
	sep := false
	for i, arg := range ah.args {
		if !sep && arg == "--" {
			sep = true
			continue
		}
		if ah.used[i] {
			continue
		}
		if !sep && len(arg) > 1 && arg[0] == '-' {
			if skipFlags {
				continue
			}
			return errs.Tag("argument error").Errorf("unknown flag: %q", arg)
		}
		if !visit(i) {
			break
		}
	}
	return nil
}

func (ah *argsHandler) PeekArgs() []string {
	out := make([]string, 0, len(ah.args))
	_ = ah.walkArgs(true, func(i int) bool {
		out = append(out, ah.args[i])
		return true
	})
	return out
}

func (ah *argsHandler) ConsumeArgs() ([]string, error) {
	out := make([]string, 0, len(ah.args))
	err := ah.walkArgs(false, func(i int) bool {
		out = append(out, ah.args[i])
		return true
	})
	if err != nil {
		return nil, err
	}
	for i := range ah.used {
		ah.used[i] = true
	}
	return out, nil
}

func (ah *argsHandler) nextArg() (int, error) {
	index := -1
	err := ah.walkArgs(false, func(i int) bool { index = i; return false })
	return index, err
}

func (ah *argsHandler) PeekArg() (string, bool, error) {
	i, err := ah.nextArg()
	if i < 0 {
		return "", false, err
	}
	return ah.args[i], true, nil
}

func (ah *argsHandler) ConsumeArg() (string, bool, error) {
	i, err := ah.nextArg()
	if i < 0 {
		return "", false, err
	}
	ah.used[i] = true
	return ah.args[i], true, nil
}

func (ah *argsHandler) ConsumeFlag(name string, bstyle bool, getenv string) (values []string, err error) {
	vals, err := ah.consumeFlags([]string{name}, bstyle)
	if err != nil || vals != nil {
		return vals, err
	}
	return ah.flagDefault(name, getenv)
}

func (ah *argsHandler) consumeFlags(names []string, bstyle bool) (values []string, err error) {
	var used []uint
	matches := func(name string) bool {
		for _, candidate := range names {
			if name == candidate {
				return true
			}
		}
		return false
	}

	for i := uint(0); i < uint(len(ah.args)); i++ {
		arg := ah.args[i]
		if ah.used[i] {
			continue
		}

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
		if idx := strings.IndexByte(arg, '='); idx >= 0 && matches(arg[:idx]) {
			values = append(values, arg[idx+1:])
			used = append(used, i)
			continue
		}

		// check if the name matches
		if !matches(arg) {
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
			return nil, errs.Tag("argument error").Errorf("no value for flag %q", arg)
		}

		// consume the next argument as the flag value
		values = append(values, ah.args[i+1])
		used = append(used, i, i+1)
		i++
	}

	for _, i := range used {
		ah.used[i] = true
	}

	return values, nil
}

func (ah *argsHandler) flagDefault(name, getenv string) (values []string, err error) {
	// if the flag was not found and we have a getenv, try
	if values == nil && getenv != "" && ah.getenv != nil {
		if val := ah.getenv(getenv); val != "" {
			values = append(values, val)
		}
	}

	// if the flag was not found, try calling the dynamic callback
	if values == nil && ah.dynamic != nil {
		return ah.dynamic(name)
	}

	return values, nil
}
