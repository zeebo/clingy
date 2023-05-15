package clingy

import (
	"strconv"
	"strings"
)

type runState struct {
	ah       *argsHandler
	pos      *paramsPos
	flags    *paramsFlags
	gflags   *paramsFlags
	names    []string
	errors   []error
	help     bool
	summary  bool
	advanced bool
}

func newRunState(name string, args []string, dynamic func(string) ([]string, error), getenv func(string) string) *runState {
	pm := newParamsMaker()
	ah := newArgsHandler(args, dynamic, getenv)

	return &runState{
		ah:     ah,
		pos:    newParamsPositional(newParamsMaker(), ah),
		flags:  newParamsFlags(pm, ah),
		gflags: newParamsFlags(pm, ah),
		names:  []string{name},
	}
}

func (st *runState) setupFlags() {
	st.help = st.gflags.Flag(
		"help", "prints help for the command", false,
		Boolean,
		Short('h'),
		Transform(strconv.ParseBool),
	).(bool)

	st.summary = st.gflags.Flag(
		"summary", "prints a summary of what commands are available", false,
		Boolean,
		Transform(strconv.ParseBool),
	).(bool)

	st.advanced = st.gflags.Flag(
		"advanced", "when used with -h, prints advanced flags help", false,
		Boolean,
		Transform(strconv.ParseBool),
	).(bool)
}

func (st *runState) params(cb func(*param)) {
	st.pos.params(cb)
	st.flags.params(cb)
	st.gflags.params(cb)
}

func (st *runState) name() string {
	return strings.Join(st.names, " ")
}

func (st *runState) peekName() (string, bool, error) {
	return st.ah.PeekArg()
}

func (st *runState) consumeName() {
	name, _, _ := st.ah.ConsumeArg() // must have been peeked
	st.names = append(st.names, name)
}

func (st *runState) hasErrors() bool {
	return st.pos.hasErrors() || st.flags.hasErrors() || st.gflags.hasErrors()
}
