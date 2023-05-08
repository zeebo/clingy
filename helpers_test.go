package clingy_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/zeebo/assert"
	"github.com/zeebo/clingy"
)

type funcCommand struct {
	SetupFn   func(params clingy.Parameters)
	ExecuteFn func(ctx context.Context) error
}

func (cmd *funcCommand) Setup(params clingy.Parameters)    { cmd.SetupFn(params) }
func (cmd *funcCommand) Execute(ctx context.Context) error { return cmd.ExecuteFn(ctx) }

func printCommand(name string) *funcCommand {
	return &funcCommand{
		SetupFn: func(params clingy.Parameters) {},
		ExecuteFn: func(ctx context.Context) error {
			fmt.Fprint(clingy.Stdout(ctx), name)
			return nil
		},
	}
}

func Run(cmd clingy.Command, args ...string) Result {
	return Capture(Env("testcommand", cmd, args...), nil)
}

func Capture(env clingy.Environment, fn func(clingy.Commands)) Result {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	env.Stdout = &stdout
	env.Stderr = &stderr
	if env.Stdin == nil {
		env.Stdin = strings.NewReader("")
	}

	ok, err := env.Run(context.Background(), fn)

	return Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Ok:     ok,
		Err:    err,
	}
}

type Result struct {
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

func (r *Result) AssertValid(t *testing.T) {
	t.Helper()
	defer logFailed(t, "stdout:\n%s", r.Stdout)
	defer logFailed(t, "stderr:\n%s", r.Stderr)
	assert.That(t, r.Ok)
	assert.NoError(t, r.Err)
}

func (r *Result) AssertStdout(t *testing.T, stdout string) {
	stdout = trimCommonSpacePrefix(stdout)
	t.Helper()
	defer logFailed(t, "stdout:\n%q", r.Stdout)
	defer logFailed(t, "expect:\n%q", stdout)
	assert.That(t, r.Stdout == stdout)
}

func (r *Result) AssertStdoutContains(t *testing.T, needle string) {
	t.Helper()
	defer logFailed(t, "stdout:\n%q", r.Stdout)
	defer logFailed(t, "needle:\n%q", needle)
	assert.That(t, strings.Contains(r.Stdout, needle))
}

func (r *Result) AssertStderr(t *testing.T, stderr string) {
	stderr = trimCommonSpacePrefix(stderr)
	t.Helper()
	defer logFailed(t, "stderr:\n%s", r.Stderr)
	defer logFailed(t, "expect:\n%s", stderr)
	assert.That(t, r.Stderr == stderr)
}

func Env(name string, root clingy.Command, args ...string) clingy.Environment {
	return clingy.Environment{
		Name: name,
		Root: root,
		Args: append([]string{}, args...), // ensure args is non-nil to avoid default
	}
}

func trimCommonSpacePrefix(x string) string {
	lines := strings.Split(x, "\n")
	for len(lines) > 0 && strings.TrimLeft(lines[0], " \t") == "" {
		lines = lines[1:]
	}
	if len(lines) == 0 {
		return ""
	}
	prefix := len(lines[0]) - len(strings.TrimLeft(lines[0], " \t"))
	for i := range lines {
		if strings.TrimSpace(lines[i]) != "" {
			lines[i] = lines[i][prefix:]
		}
	}
	return strings.TrimRight(strings.Join(lines, "\n"), " \t")
}
