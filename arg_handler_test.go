package clingy

import (
	"errors"
	"testing"

	"github.com/zeebo/assert"
	"github.com/zeebo/errs/v2"
)

func TestArgHandler(t *testing.T) {
	ah := newArgsHandler([]string{
		"--foo", "bar", // flag
		"baz",            // arg
		"--bif", "--baf", // flag
		"--zap=true",    // flag
		"arg",           // arg
		"--extra",       // flag
		"--zap=false",   // flag
		"--zap",         // flag
		"--",            // separator
		"arg2",          // arg
		"--foo", "bing", // args
	}, func(name string) ([]string, error) {
		if name == "err" {
			return nil, errs.Tag("sentinel")
		}
		return []string{"sym"}, nil
	}, func(name string) string {
		if name == "ENV_ENV" {
			return "envval"
		}
		return ""
	})

	{ // first peek doesn't know if "bar" is the value for "--foo" or if "--foo" is boolean
		assert.DeepEqual(t, ah.PeekArgs(), []string{"bar", "baz", "arg", "arg2", "--foo", "bing"})
	}

	{ // can't consume the first arg because it is a flag
		_, _, err := ah.ConsumeArg()
		assert.That(t, errors.Is(err, errs.Tag("argument error")))
	}

	{ // parse "--foo", "bar" is removed from args
		got, err := ah.ConsumeFlag("foo", false, "ENV_FOO")
		assert.NoError(t, err)
		assert.DeepEqual(t, got, []string{"bar"})
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "arg", "arg2", "--foo", "bing"})
	}

	{ // if "--zap" is not boolean, the final "--zap" has no value associated
		got, err := ah.ConsumeFlag("zap", false, "ENV_ZAP")
		assert.That(t, errors.Is(err, errs.Tag("argument error")))
		assert.Nil(t, got)
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "arg", "arg2", "--foo", "bing"})
	}

	{ // parse "--zap" as boolean, getting 3 values
		got, err := ah.ConsumeFlag("zap", true, "ENV_ZAP")
		assert.NoError(t, err)
		assert.DeepEqual(t, got, []string{"true", "false", "true"})
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "arg", "arg2", "--foo", "bing"})
	}

	{ // there is no "--baf" flag because it is a potential value to "--bif", so this is an error
		got, err := ah.ConsumeFlag("baf", false, "ENV_BAF")
		assert.That(t, errors.Is(err, errs.Tag("argument error")))
		assert.Nil(t, got)
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "arg", "arg2", "--foo", "bing"})
	}

	{ // parse "--bif" consuming the "--baf" value
		got, err := ah.ConsumeFlag("bif", false, "ENV_BIF")
		assert.NoError(t, err)
		assert.DeepEqual(t, got, []string{"--baf"})
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "arg", "arg2", "--foo", "bing"})
	}

	{ // ensure that the dynamic callback can be used to successfully return a value
		got, err := ah.ConsumeFlag("not-exist", false, "ENV_NOT_EXIST")
		assert.NoError(t, err)
		assert.DeepEqual(t, got, []string{"sym"})
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "arg", "arg2", "--foo", "bing"})
	}

	{ // ensure that the dynamic callback can be used to return an error
		got, err := ah.ConsumeFlag("err", false, "ENV_ERR")
		assert.That(t, errors.Is(err, errs.Tag("sentinel")))
		assert.Nil(t, got)
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "arg", "arg2", "--foo", "bing"})
	}

	{ // ensure that the environment callback can be used to parse a value
		got, err := ah.ConsumeFlag("env", false, "ENV_ENV")
		assert.NoError(t, err)
		assert.DeepEqual(t, got, []string{"envval"})
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "arg", "arg2", "--foo", "bing"})
	}

	{ // consume the first "baz" argument
		got, ok, err := ah.ConsumeArg()
		assert.NoError(t, err)
		assert.That(t, ok)
		assert.DeepEqual(t, got, "baz")
		assert.DeepEqual(t, ah.PeekArgs(), []string{"arg", "arg2", "--foo", "bing"})
	}

	{ // attempting to consume all the remaining args errors as there is still a potential flag
		_, err := ah.ConsumeArgs()
		assert.That(t, errors.Is(err, errs.Tag("argument error")))
	}

	{ // consume the remaining extra flag
		got, err := ah.ConsumeFlag("extra", true, "ENV_EXTRA")
		assert.NoError(t, err)
		assert.DeepEqual(t, got, []string{"true"})
		assert.DeepEqual(t, ah.PeekArgs(), []string{"arg", "arg2", "--foo", "bing"})
	}

	{ // consume the first "arg" argument
		got, ok, err := ah.ConsumeArg()
		assert.NoError(t, err)
		assert.That(t, ok)
		assert.DeepEqual(t, got, "arg")
		assert.DeepEqual(t, ah.PeekArgs(), []string{"arg2", "--foo", "bing"})
	}

	{ // peek the argument that is right after the "--""
		got, ok, err := ah.PeekArg()
		assert.NoError(t, err)
		assert.That(t, ok)
		assert.DeepEqual(t, got, "arg2")
		assert.DeepEqual(t, ah.PeekArgs(), []string{"arg2", "--foo", "bing"})
	}

	{ // consume the second "arg2" argument
		got, ok, err := ah.ConsumeArg()
		assert.NoError(t, err)
		assert.That(t, ok)
		assert.DeepEqual(t, got, "arg2")
		assert.DeepEqual(t, ah.PeekArgs(), []string{"--foo", "bing"})
	}

	{ // no flags means we can consume the rest of the args for a repeated arg
		got, err := ah.ConsumeArgs()
		assert.NoError(t, err)
		assert.DeepEqual(t, got, []string{"--foo", "bing"})
		assert.DeepEqual(t, ah.PeekArgs(), []string{})
	}
}
