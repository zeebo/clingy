package clingy

import (
	"fmt"
	"io"
)

type headerPrinter struct {
	hdr  string
	done bool
	w    io.Writer
}

func newHeaderPrinter(w io.Writer, hdr string) *headerPrinter {
	return &headerPrinter{w: w, hdr: hdr}
}

func (h *headerPrinter) Write(p []byte) (n int, err error) {
	if !h.done {
		fmt.Fprintln(h.w, "\n"+h.hdr)
		h.done = true
	}
	return h.w.Write(p)
}
