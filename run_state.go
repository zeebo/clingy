package clingy

import (
	"strconv"
	"strings"
)

type runState struct {
	pos      *paramsPos
	flags    *paramsFlags
	gflags   *paramsFlags
	names    []string
	errors   []error
	help     bool
	advanced bool
}

func newRunState(name string, args []string, dynamic func(string) ([]string, error)) *runState {
	pos := make([]string, 0)
	flags := make(map[string][]string)
	allPositional := false

	for i := uint(0); i < uint(len(args)); i++ {
		arg := args[i]

		// check if the argument ends all flags
		if arg == "--" {
			allPositional = true
			continue
		}

		// check if the argument is positional
		if len(arg) < 1 || arg[0] != '-' || arg == "-" || allPositional {
			pos = append(pos, arg)
			continue
		}

		// strip off the -- prefix for easier processing
		if len(arg) >= 2 && arg[:2] == "--" {
			arg = arg[2:]
		} else {
			arg = arg[1:]
		}

		// check for --foo=bar form
		if idx := strings.IndexByte(arg, '='); idx >= 0 {
			flags[arg[:idx]] = append(flags[arg[:idx]], arg[idx+1:])
			continue
		}

		// check if the flag is the final argument
		if i+1 >= uint(len(args)) {
			flags[arg] = append(flags[arg], "true")
			continue
		}

		// check if the next argument is also a flag
		value := args[i+1]
		if len(value) >= 1 && value[0] == '-' {
			flags[arg] = append(flags[arg], "true")
			continue
		}

		// consume the next argument as a flag
		flags[arg] = append(flags[arg], value)
		i++
	}

	return &runState{
		pos:    newParamsPositional(pos),
		flags:  newParamsFlags(flags, nil),
		gflags: newParamsFlags(flags, dynamic),
		names:  []string{name},
	}
}

func (st *runState) setupFlags() {
	st.help = st.gflags.New(
		"help", "prints help for the command", false,
		Short('h'), Transform(strconv.ParseBool)).(bool)
	st.advanced = st.gflags.New(
		"advanced", "when used with -h, prints advanced flags help", false,
		Transform(strconv.ParseBool)).(bool)
}

func (st *runState) params(cb func(*param)) {
	st.pos.params(cb)
	st.flags.params(cb)
	st.gflags.params(cb)
}

func (st *runState) name() string {
	return strings.Join(st.names, " ")
}

func (st *runState) firstPos() (string, bool) {
	if len(st.pos.pos) == 0 {
		return "", false
	}
	return st.pos.pos[0], true
}

func (st *runState) pushName() {
	st.names = append(st.names, st.pos.pos[0])
	st.pos.pos = st.pos.pos[1:]
}

func (st *runState) hasErrors() bool {
	return st.pos.hasErrors() || st.flags.hasErrors() || st.gflags.hasErrors()
}
