// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package clingy

import (
	"context"
	"io"
)

type stdioKeyType string

const stdioKey stdioKeyType = "environment"

type stdioEnvironment struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

// Stdin returns the io.Reader from the Environment associated to the context.
func Stdin(ctx context.Context) io.Reader {
	val, _ := ctx.Value(stdioKey).(stdioEnvironment)
	return val.stdin
}

// Stdout returns the io.Writer from the Environment associated to the context.
func Stdout(ctx context.Context) io.Writer {
	val, _ := ctx.Value(stdioKey).(stdioEnvironment)
	return val.stdout
}

// Stderr returns the io.Writer from the Environment associated to the context.
func Stderr(ctx context.Context) io.Writer {
	val, _ := ctx.Value(stdioKey).(stdioEnvironment)
	return val.stderr
}
