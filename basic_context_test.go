package clingy

import (
	"context"
	"testing"
)

func TestBasicStdioWrapping(t *testing.T) {
	ctx := context.Background()
	bctx := WithBufferedStdio(ctx)
	err := doSomething(bctx)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if bctx.GetWrittenOut() != "Just some output" {
		t.Fatalf("Invalid output: %s", bctx.GetWrittenOut())
	}
}

func doSomething(ctx context.Context) error {
	bctx := NewBasicContext(ctx)
	_, err := bctx.Write([]byte("Just some output"))
	return err
}
