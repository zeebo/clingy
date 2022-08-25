package clingy

import (
	"bytes"
	"context"
	"testing"
)

func TestBasicStdioWrapping(t *testing.T) {
	ctx := context.Background()
	ctx = WithBufferedStdio(ctx)
	err := doSomething(ctx)
	if err != nil {
		t.Fatalf("%v", err)
	}
	out := Stdout(ctx).(*bytes.Buffer)
	if out.String() != "Just some output" {
		t.Fatalf("Invalid output: %s", out.String())
	}
}

func doSomething(ctx context.Context) error {
	_, err := Stdout(ctx).Write([]byte("Just some output"))
	return err
}
