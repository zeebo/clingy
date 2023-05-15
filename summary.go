package clingy

import (
	"context"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

func (env *Environment) printSummary(ctx context.Context, st *runState, desc cmdDesc) {
	tw := tabwriter.NewWriter(env.Stdout, 4, 4, 4, ' ', 0)
	defer tw.Flush()

	fmt.Fprintln(tw, "Available commands:")
	printSubcommandsRecursive(ctx, tw, st.names, desc)
}

func printSubcommandsRecursive(ctx context.Context, w io.Writer, name []string, desc cmdDesc) {
	for _, desc := range desc.subcmds {
		dname := append(name, desc.name)
		if desc.cmd != nil {
			fmt.Fprintf(w, "\t%s\t%s\n", strings.Join(dname, " "), desc.short)
		}
		printSubcommandsRecursive(ctx, w, dname, desc)
	}
}
