package clingy

import (
	"context"
	"os"

	"github.com/zeebo/errs/v2"
)

// Run calls the fn to create and execute the tree of commands and global flags.
// It returns a boolean indicating if the parsing/dispatching of the command
// was successful. The error is the returned error from any executed command.
func (env Environment) Run(ctx context.Context, fn func(Commands)) (bool, error) {
	if env.Name == "" {
		env.Name = os.Args[0]
	}
	if env.Args == nil {
		env.Args = os.Args[1:]
	}
	if env.Stdin == nil {
		env.Stdin = os.Stdin
	}
	if env.Stdout == nil {
		env.Stdout = os.Stdout
	}
	if env.Stderr == nil {
		env.Stderr = os.Stderr
	}

	st := newRunState(env.Name, env.Args, env.Dynamic)
	descs := collectDescs(st.gflags, fn)
	st.setupFlags()

	valid, usage, err := env.dispatchDesc(ctx, st, cmdDesc{
		cmd:     env.Root,
		subcmds: descs,
	})
	if !usage {
		env.appendUnknownCommandErrorWithSuggestions(st, descs)
		env.printUsage(ctx, st, cmdDesc{subcmds: descs})
	}
	return valid, err
}

func (env *Environment) dispatch(ctx context.Context, st *runState, descs []cmdDesc) (bool, bool, error) {
	name, ok, err := st.peekName()
	if err != nil || !ok {
		return false, false, err
	}

	for _, desc := range descs {
		if desc.name != name {
			continue
		}
		st.consumeName()
		return env.dispatchDesc(ctx, st, desc)
	}

	return false, false, nil
}

func (env *Environment) dispatchDesc(ctx context.Context, st *runState, desc cmdDesc) (valid bool, usage bool, err error) {
	if valid, usage, err := env.dispatch(ctx, st, desc.subcmds); valid {
		return valid, usage, err
	}

	if desc.cmd != nil {
		desc.cmd.Setup(newParams(st.pos, st.flags))
	}

	// handle any errors parsing the arguments
	if st.hasErrors() {
		if !st.help {
			st.params(func(p *param) {
				if p == nil {
					return
				}
				if p.err != nil {
					st.errors = append(st.errors, errs.Tag("argument error").Wrap(p.err))
				}
			})
		}
		env.printUsage(ctx, st, desc)
		return true, true, nil
	}

	// handle any dynamic errors surfacing from setting up flags by returning the error
	// directly up through the Run call.
	if err := st.flags.err; err != nil {
		return true, false, err
	} else if err := st.gflags.err; err != nil {
		return true, false, err
	}

	// if we don't have a command to execute, check if it's because they
	// specified the wrong name, and error if so.
	if desc.cmd == nil {
		if len(desc.subcmds) > 0 {
			env.appendUnknownCommandErrorWithSuggestions(st, desc.subcmds)
		}
		env.printUsage(ctx, st, desc)
		return false, true, nil
	}

	// print usage if requested
	if st.help {
		env.printUsage(ctx, st, desc)
		return true, true, nil
	}

	ctx = context.WithValue(ctx, stdioKey, stdioEnvironment{
		stdin:  env.Stdin,
		stdout: env.Stdout,
		stderr: env.Stderr,
	})

	if env.Wrap != nil {
		err = env.Wrap(ctx, desc.cmd)
	} else {
		err = desc.cmd.Execute(ctx)
	}
	return true, true, err
}
