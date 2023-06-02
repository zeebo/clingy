package clingy

import (
	"context"
	"os"
	"path/filepath"

	"github.com/zeebo/errs/v2"
)

func (env *Environment) fillDefaults() {
	if env.Name == "" {
		env.Name = filepath.Base(os.Args[0])
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
	if env.Getenv == nil {
		env.Getenv = os.Getenv
	}
}

// Run calls the fn to create and execute the tree of commands and global flags.
// It returns a boolean indicating if the parsing/dispatching of the command
// was successful. The error is the returned error from any executed command.
func (env Environment) Run(ctx context.Context, fn func(Commands)) (bool, error) {
	env.fillDefaults()
	st := newRunState(env.Name, env.Args, env.Dynamic, env.Getenv)
	descs := collectDescs(st.gflags, fn)
	st.setupFlags()

	executed, _, err := env.dispatchDesc(ctx, st, cmdDesc{
		cmd:     env.Root,
		subcmds: descs,
	})
	return executed, err
}

func (env *Environment) dispatch(ctx context.Context, st *runState, descs []cmdDesc) (executed bool, matched bool, err error) {
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

func (env *Environment) dispatchDesc(ctx context.Context, st *runState, desc cmdDesc) (executed bool, matched bool, err error) {
	if executed, matched, err := env.dispatch(ctx, st, desc.subcmds); matched {
		return executed, matched, err
	}

	if desc.cmd != nil {
		desc.cmd.Setup(newParams(st.pos, st.flags))
	}

	// print usage if requested
	if st.help {
		env.printUsage(ctx, st, desc)
		return true, true, nil
	}

	// print summary if required
	if st.summary {
		env.printSummary(ctx, st, desc)
		return true, true, nil
	}

	// handle any errors parsing the arguments
	if st.hasErrors() {
		if !st.help {
			st.params(func(p *param) {
				if p != nil && p.err != nil {
					st.errors = append(st.errors, errs.Tag("argument error").Wrap(p.err))
				}
			})
		}
		env.printUsage(ctx, st, desc)
		return false, true, nil
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

	// consume the remaining arguments. if there are any, error.
	if args, err := st.ah.ConsumeArgs(); err != nil || len(args) > 0 {
		if err != nil {
			st.errors = append(st.errors, err)
		} else {
			st.errors = append(st.errors, errs.Tag("argument error").Errorf("unknown arguments: %q", args))
		}
		env.printUsage(ctx, st, desc)
		return false, true, nil
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
