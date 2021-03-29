package clingy

import (
	"context"
	"io"
)

type basicContext struct {
	context.Context

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func (b basicContext) WithContext(ctx context.Context) Context {
	b.Context = ctx
	return b
}

func (b basicContext) Read(p []byte) (n int, err error)  { return b.stdin.Read(p) }
func (b basicContext) Write(p []byte) (n int, err error) { return b.stdout.Write(p) }

func (b basicContext) Stdin() io.Reader  { return b.stdin }
func (b basicContext) Stdout() io.Writer { return b.stdout }
func (b basicContext) Stderr() io.Writer { return b.stderr }
