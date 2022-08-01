// Package clingy is a library to create command line interfaces.
//
// The API is a design experiment that prioritizes discoverability
// of the code that uses it over anything else. There are two main
// ways that this can be seen.
//
// First, commands are defined in the same call that is used to
// execute them. In this way, you cannot define the tree of commands
// in a bunch of different files in the package, mutating state
// during init() functions, or otherwise. For example, consider
// the following execution:
//
//	ok, err := env.Run(ctx, func(cmds clingy.Commands) {
//		cmds.Group("files", "Commands related to files", func() {
//			cmds.New("copy", "Copy a file", new(cmdFilesCopy))
//			cmds.New("delete", "Delete a file", new(cmdFilesDelete))
//			cmds.New("read", "Read a file", new(cmdFilesRead))
//			cmds.New("list", "List some files", new(cmdFilesList))
//		})
//		...
//
// Because all of the commands are defined in that single spot, we know
// exactly which commands are available and where they are defined.
//
// Secondly, flags and arguments for commands are defined at the time
// they are loaded, and happen as part of a method call on the command
// implementation itself. This provides incentives to store the state
// on the command object itself. For example:
//
//	type someCommand struct {
//		first  string
//		second string
//	}
//
//	func (s *someCommand) Setup(params clingy.Parameters) {
//		s.first = params.Arg("first", "first required argument").(string)
//		s.second = params.Arg("second", "second required argument").(string)
//	}
//
// This causes the command implementation to be more discoverable because
// you have a single location that contains all of the state the command
// will inspect. This also aids in unit testing, because you can construct
// values directly and execute them, skipping the Setup step and avoiding
// having to make mocks or fakes to inject the parameters.
//
// In service of this goal, some tradeoffs are made. Most notably, the return
// type for arguments and flags is `interface{}` and type casts are required.
// This was found to be acceptable because in typical usage, the argument will
// be directly assigned into a typed struct field, so the compiler complains
// if an incorrect type assertion is used. This changes a dynamic failure to
// a minor redundancy. Additionally, the code with the type assertions is
// executed right away whenever a command is run, so any testing will surface
// any bugs.
//
// Enjoy!
package clingy

import (
	"context"
	"io"
)

// Command is the interface that executable commands implement.
type Command interface {
	// Setup is called to define the positional arguments and flags for the command.
	// The returned values should be stored for the upcoming call to Execute.
	Setup(params Parameters)

	// Execute is called after Setup and should run the command.
	Execute(ctx context.Context) error
}

// Option is the type for values that control details around argument and flags like
// their presentation, if they are repeated or optional, etc.
type Option struct {
	do func(*paramOpts)
}

var (
	// Repeated sets the flag or argument to be repeated, returning a slice of values.
	// Repeated arguments must come after any other arguments, Optional or otherwise.
	// If not, New will panic.
	Repeated = Option{func(po *paramOpts) { po.rep = true }}

	// Optional sets the argument to be optional, returning a pointer to a value.
	// Optional arguments must come after any required arguments and before any
	// Repeated arguments. If not, New will panic.
	Optional = Option{func(po *paramOpts) { po.opt = true }}

	// Advanced causes the flag to be hidden unless the --advanced flag is specified
	// when usage information is printed.
	Advanced = Option{func(po *paramOpts) { po.adv = true }}

	// Boolean causes the flag to be considered a "boolean style" flag where it does
	// not look at the next positional argument if no value is specified.
	Boolean = Option{func(po *paramOpts) { po.bstyle = true }}
)

// Short causes the flag to be able to be specified with a single character.
func Short(c byte) Option {
	return Option{func(po *paramOpts) { po.short = c }}
}

// Transform takes a list of functions meant to parse and transform a string into some
// final result type. The functions must be of the form (borrowing generics syntax)
//
//    func[type T, S any](x T) (y S, err error)
//
// The first function always has an input of type string. The functions are called
// in sequence. If the argument or flag is Repeated, the functions are called on each
// element of the values. If Transform is specified multiple times, the functions
// are appended. In other words, the following two calls are equivalent:
//
//    args.New(..., Transform(f1), Transform(f2))
//    args.New(..., Transform(f1, f2))
func Transform(fns ...interface{}) Option {
	return Option{func(po *paramOpts) { po.fns = append(po.fns, fns...) }}
}

// Type specifies what type to show in the usage for a flag. For example, specifying
//
//     type MyInt int
//     args.New("foo", "some foo flag", MyInt(5), Type("my_int"), Transform(..)).(MyInt)
//
// would have a usage that looks like
//
//   --foo my_int    some foo flag (default 5)
func Type(typ string) Option {
	return Option{func(po *paramOpts) { po.typ = typ }}
}

type Parameters interface {
	// Flags is embedded to allow one to create command level flags.
	Flags

	// Arg creates a new argument. The return value is the value of the argument.
	// If the Repeated option is specified, then the return type is a slice of
	// whatever it would have been. Otherwise, if the Optional option is specified,
	// the return type is a pointer to whatever it would have been.
	//
	// Arg panics if the same name is defined twice. Arg panics if any arguments
	// are created after a Repeated argument is created. Arg panics if any arguments
	// that are not Optional or Repeated are created after an Optional argument is
	// created.
	Arg(name, desc string, options ...Option) interface{}
}

// Flags allows the creation of flags as well as retreiving their values.
type Flags interface {
	// Flag creates a new flag. The return value is the value of the flag.
	// If the Repeated option is specified, then the return type is a slice of
	// whatever it would have been. Optional has no effect. The value provided
	// in def is returned if the flag was not specified.
	//
	// Flag panics if the same name is defined twice, or if the same Short option
	// is used twice.
	Flag(name, desc string, def interface{}, options ...Option) interface{}

	// Break inserts a line break in the usage output of the flags.
	Break()
}

// Commands is used to construct the tree of commands and subcommands.
type Commands interface {
	// Flags is embedded to allow one to create global flags.
	Flags

	// New creates a new command.
	New(name, desc string, cmd Command)

	// Group begins a new command group. Calls to New inside of the children
	// function are associated with the most recent call to Group.
	Group(name, desc string, children func())
}

// Environment is used to control which command is run, what flags and arguments
// it receives, and what input/output it has access to.
type Environment struct {
	// Name is the name of the binary being executed.
	// If empty, os.Args[0] is used.
	Name string

	// Args is consulted to determine values for the flags and positional arguments.
	// If empty, os.Args[1:] is used.
	Args []string

	// Dynamic, if set, is consulted for global flag values if they are not
	// specified as part of Args. If an error is returned, it will no longer
	// be consulted, and the error will be returned from Run.
	Dynamic func(name string) (vals []string, err error)

	// Wrap, if set, is called with the context and command that would have
	// been executed. The no-op implementation is `return cmd.Execute(ctx)`.
	Wrap func(ctx context.Context, cmd Command) (err error)

	// SuggestionsMinEditDistance defines minimum Levenshtein distance to
	// display suggestions when a command/subcommand is misspelled.
	// Must be > 0.
	// Default is 2
	SuggestionsMinEditDistance int

	// 	DisableSuggestions disables the suggestions based on Levenshtein
	//	distance that go along with 'unknown command' messages
	DisableSuggestions bool

	Stdin  io.Reader // Stdin defaults to os.Stdin if unset.
	Stdout io.Writer // Stdout defaults to os.Stdout if unset.
	Stderr io.Writer // Stderr defaults to os.Stderr if unset.
}
