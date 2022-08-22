// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package clingy

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/zeebo/errs/v2"
)

var stdinKey, stdoutKey, stderrKey struct{}

// StdioContext can help to test CLI apps with wrapping stdout/stdin/stderr.
type StdioContext struct {
	context.Context
}

// StdioTestContext is for testing, where stdio is replaced with in-memory buffers.
type StdioTestContext struct {
	StdioContext
}

func WrapStdio(ctx context.Context) StdioContext {
	ctx = context.WithValue(ctx, stdinKey, os.Stdin)
	ctx = context.WithValue(ctx, stderrKey, os.Stderr)
	ctx = context.WithValue(ctx, stdoutKey, os.Stdout)
	return StdioContext{
		Context: ctx,
	}
}

// WithBufferedStdio creates context with in-memory stdio targets.
func WithBufferedStdio(ctx context.Context) StdioTestContext {
	ctx = context.WithValue(ctx, stdinKey, &bytes.Buffer{})
	ctx = context.WithValue(ctx, stderrKey, &bytes.Buffer{})
	ctx = context.WithValue(ctx, stdoutKey, &bytes.Buffer{})
	return StdioTestContext{
		StdioContext{
			ctx,
		},
	}
}

func NewBasicContext(ctx context.Context) StdioContext {
	return StdioContext{
		Context: ctx,
	}
}

func (b StdioContext) Read(p []byte) (n int, err error) {
	if b.Value(stdinKey) == nil {
		return 0, errs.Errorf("stdin is not wrapped")
	}
	return b.Stdin().Read(p)
}
func (b StdioContext) Write(p []byte) (n int, err error) {
	if b.Value(stdoutKey) == nil {
		return 0, errs.Errorf("stdout is not wrapped")
	}
	return b.Stdout().Write(p)
}

func (b StdioContext) Stdin() io.Reader {
	return b.Value(stdinKey).(io.Reader)
}
func (b StdioContext) Stdout() io.Writer {
	return b.Value(stdoutKey).(io.Writer)
}
func (b StdioContext) Stderr() io.Writer {
	return b.Value(stderrKey).(io.Writer)
}

func (b StdioTestContext) GetWrittenOut() string {
	return string(b.Value(stdoutKey).(*bytes.Buffer).Bytes())
}

func (b StdioTestContext) GetWrittenErr() string {
	return string(b.Value(stdoutKey).(*bytes.Buffer).Bytes())
}
