package clingy

import (
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
		"--zap=false",   // flag
		"--zap",         // flag
		"--",            // separator
		"--foo", "bing", // args
	}, func(name string) ([]string, error) {
		if name == "err" {
			return nil, errs.Errorf("sentinel")
		}
		return []string{"sym"}, nil
	}, func(name string) string {
		if name == "ENV_ENV" {
			return "envval"
		}
		return ""
	})

	{
		got, err := ah.ConsumeFlag("foo", false, "ENV_FOO")
		assert.NoError(t, err)
		assert.DeepEqual(t, got, []string{"bar"})
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "--bif", "--baf", "--zap=true", "arg", "--zap=false", "--zap", "--foo", "bing"})
	}

	{
		got, err := ah.ConsumeFlag("zap", false, "ENV_ZAP")
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "--bif", "--baf", "--zap=true", "arg", "--zap=false", "--zap", "--foo", "bing"})
	}

	{
		got, err := ah.ConsumeFlag("zap", true, "ENV_ZAP")
		assert.NoError(t, err)
		assert.DeepEqual(t, got, []string{"true", "false", "true"})
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "--bif", "--baf", "arg", "--foo", "bing"})
	}

	{
		got, err := ah.ConsumeFlag("baf", false, "ENV_BAF")
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "--bif", "--baf", "arg", "--foo", "bing"})
	}

	{
		got, err := ah.ConsumeFlag("bif", false, "ENV_BIF")
		assert.NoError(t, err)
		assert.DeepEqual(t, got, []string{"--baf"})
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "arg", "--foo", "bing"})
	}

	{
		got, err := ah.ConsumeFlag("not-exist", false, "ENV_NOT_EXIST")
		assert.NoError(t, err)
		assert.DeepEqual(t, got, []string{"sym"})
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "arg", "--foo", "bing"})
	}

	{
		got, err := ah.ConsumeFlag("err", false, "ENV_ERR")
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "arg", "--foo", "bing"})
	}

	{
		got, err := ah.ConsumeFlag("env", false, "ENV_ENV")
		assert.NoError(t, err)
		assert.DeepEqual(t, got, []string{"envval"})
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "arg", "--foo", "bing"})
	}

	{
		got, ok, err := ah.ConsumeArg("first")
		assert.NoError(t, err)
		assert.That(t, ok)
		assert.DeepEqual(t, got, "baz")
		assert.DeepEqual(t, ah.PeekArgs(), []string{"arg", "--foo", "bing"})
	}

	{
		got, err := ah.ConsumeArgs()
		assert.NoError(t, err)
		assert.DeepEqual(t, got, []string{"arg", "--foo", "bing"})
		assert.DeepEqual(t, ah.PeekArgs(), []string{})
	}
}
