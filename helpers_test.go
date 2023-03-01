package clingy

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/zeebo/assert"
)

//
// Capture records the execution of some commands in some environment
//

func Capture(env Environment, fn func(*RecordingCmds)) Result {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	var rc *RecordingCmds

	env.Stdout = stdout
	env.Stderr = stderr

	ok, err := env.Run(context.Background(), func(cmds Commands) {
		rc = newRecordingCmds(cmds)
		fn(rc)
	})

	return Result{
		Cmds:   rc.Saved(),
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Ok:     ok,
		Err:    err,
	}
}

func Env(name string, args ...string) Environment {
	return Environment{
		Name: name,
		// ensure args is non-nil to avoid default
		Args: append([]string{}, args...),
	}
}

//
// Result keeps track of the results of a Environment.Run call and
// has helper assertion methods to make tests concise.
//

type Result struct {
	Cmds   map[string]*RecordingCmd
	Stdout string
	Stderr string
	Ok     bool
	Err    error
}

func logFailed(t *testing.T, format string, args ...interface{}) {
	if t.Failed() {
		t.Logf(format, args...)
	}
}

func (r *Result) AssertRunValid(t *testing.T) {
	t.Helper()
	defer logFailed(t, "stdout:\n%s", r.Stdout)
	defer logFailed(t, "stderr:\n%s", r.Stderr)
	assert.That(t, r.Ok)
	assert.NoError(t, r.Err)
}

func (r *Result) AssertExecuted(t *testing.T, name string) {
	t.Helper()
	for cname, cmd := range r.Cmds {
		cmd.AssertExecution(t, name == cname)
	}
}

func (r *Result) AssertStdoutContains(t *testing.T, needle string) {
	t.Helper()
	defer logFailed(t, "stdout:\n%s", r.Stdout)
	assert.That(t, strings.Contains(r.Stdout, needle))
}

func (r *Result) AssertStderrContains(t *testing.T, needle string) {
	t.Helper()
	defer logFailed(t, "stderr:\n%s", r.Stderr)
	assert.That(t, strings.Contains(r.Stderr, needle))
}

//
// recordingCmds keeps track of all of the commands defined and executed
// during a test.
//

type RecordingCmds struct {
	stack []string
	saved map[string]*RecordingCmd
	cmds  Commands
}

func newRecordingCmds(cmds Commands) *RecordingCmds {
	return &RecordingCmds{
		saved: make(map[string]*RecordingCmd),
		cmds:  cmds,
	}
}

func (r *RecordingCmds) Saved() map[string]*RecordingCmd {
	return r.saved
}

func (r *RecordingCmds) Flag(name, desc string, def interface{}, options ...Option) interface{} {
	return r.cmds.Flag(name, desc, def, options...)
}

func (r *RecordingCmds) Break() {
	r.cmds.Break()
}

func (r *RecordingCmds) New(name string) {
	fullName := strings.Join(append(r.stack, name), " ")
	cmd := NewRecordingCmd(fullName)
	r.cmds.New(name, name+" command", cmd)
	r.saved[fullName] = cmd
}

func (r *RecordingCmds) Group(name string, children func()) {
	r.stack = append(r.stack, name)
	r.cmds.Group(name, name+" group", children)
	r.stack = r.stack[:len(r.stack)-1]
}

//
// recordingCmd is a command that defines a bunch of settings and keeps
// track of if it was called, etc.
//

type RecordingCmd struct {
	name     string
	setup    bool
	executed bool

	Flags RecordingFlags
	Args  RecordingArgs
}

type RecordingFlags struct {
	ValString string
	ValInt    int
	ValBool   bool

	OptString *string
	OptInt    *int
	OptBool   *bool

	RepString []string
	RepInt    []int
	RepBool   []bool

	HiddenString string
	HiddenInt    int
	HiddenBool   bool
}

func (rf *RecordingFlags) Setup(flags Flags) {
	parseInt := Transform(strconv.Atoi)
	parseBool := Transform(strconv.ParseBool)

	rf.ValString = flags.Flag("Flags.ValString", "", nil).(string)
	rf.ValInt = flags.Flag("Flags.ValInt", "", nil, parseInt).(int)
	rf.ValBool = flags.Flag("Flags.ValBool", "", nil, Boolean, parseBool).(bool)

	rf.OptString = flags.Flag("Flags.OptString", "", nil, Optional).(*string)
	rf.OptInt = flags.Flag("Flags.OptInt", "", nil, Optional, parseInt).(*int)
	rf.OptBool = flags.Flag("Flags.OptBool", "", nil, Optional, Boolean, parseBool).(*bool)

	rf.RepString = flags.Flag("Flags.RepString", "", nil, Repeated).([]string)
	rf.RepInt = flags.Flag("Flags.RepInt", "", nil, Repeated, parseInt).([]int)
	rf.RepBool = flags.Flag("Flags.RepBool", "", nil, Repeated, Boolean, parseBool).([]bool)

	rf.HiddenString = flags.Flag("Flags.HiddenString", "", nil, Hidden).(string)
	rf.HiddenInt = flags.Flag("Flags.HiddenInt", "", nil, parseInt, Hidden).(int)
	rf.HiddenBool = flags.Flag("Flags.HiddenBool", "", nil, Boolean, parseBool, Hidden).(bool)
}

type RecordingArgs struct {
	ValString string
	ValInt    int

	OptString *string
	OptInt    *int

	RepInt []int
}

func (ra *RecordingArgs) Setup(params Parameters) {
	parseInt := Transform(strconv.Atoi)

	ra.ValString = params.Arg("Args.ValString", "").(string)
	ra.ValInt = params.Arg("Args.ValInt", "", parseInt).(int)

	ra.OptString = params.Arg("Args.OptString", "", Optional).(*string)
	ra.OptInt = params.Arg("Args.OptInt", "", Optional, parseInt).(*int)

	ra.RepInt = params.Arg("Args.RepInt", "", Repeated, parseInt).([]int)
}

func NewRecordingCmd(name string) *RecordingCmd {
	return &RecordingCmd{
		name: name,
	}
}

func (r *RecordingCmd) Setup(params Parameters) {
	r.setup = true

	r.Flags.Setup(params)
	r.Args.Setup(params)
}

func (r *RecordingCmd) Execute(ctx context.Context) error {
	r.executed = true
	return nil
}

func (r *RecordingCmd) AssertExecution(t *testing.T, executed bool) {
	t.Helper()
	if executed {
		r.AssertExecuted(t)
	} else {
		r.AssertNotExecuted(t)
	}
}

func (r *RecordingCmd) AssertExecuted(t *testing.T) {
	t.Helper()
	defer logFailed(t, "command %q unexpectedly didn't execute", r.name)
	assert.That(t, r.setup)
	assert.That(t, r.executed)
}

func (r *RecordingCmd) AssertNotExecuted(t *testing.T) {
	t.Helper()
	defer logFailed(t, "command %q unexpectedly executed", r.name)
	assert.That(t, !r.setup)
	assert.That(t, !r.executed)
}
