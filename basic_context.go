// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package clingy

import (
	"bytes"
	"context"
	"io"
	"os"
)

type stdioContextKey string

const stdioKey stdioContextKey = "environment"

type stdioEnvironment struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

// WrapStdio saves stdin/out/err to the context for later use.
func WrapStdio(ctx context.Context) context.Context {
	return context.WithValue(ctx, stdioKey, stdioEnvironment{
		stdout: os.Stdout,
		stdin:  os.Stdin,
		stderr: os.Stderr,
	})
}

// WithBufferedStdio creates context with in-memory stdio targets.
func WithBufferedStdio(ctx context.Context) context.Context {
	return context.WithValue(ctx, stdioKey, stdioEnvironment{
		stdout: &bytes.Buffer{},
		stdin:  &bytes.Buffer{},
		stderr: &bytes.Buffer{},
	})
}

func Stdin(ctx context.Context) io.Reader {
	return ctx.Value(stdioKey).(stdioEnvironment).stdin
}

func Stdout(ctx context.Context) io.Writer {
	return ctx.Value(stdioKey).(stdioEnvironment).stdout
}

func Stderr(ctx context.Context) io.Writer {
	return ctx.Value(stdioKey).(stdioEnvironment).stderr
}
