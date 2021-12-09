package clingy

import (
	"testing"

	"github.com/zeebo/assert"
	"github.com/zeebo/errs/v2"
)

func TestArgHandler(t *testing.T) {
	ah := newArgsHandler([]string{
		"--foo", "bar", "baz", "--bif", "--baf", "--zap=true", "arg", "--zap=false", "--zap", "--", "--foo", "bing",
	}, func(name string) ([]string, error) {
		if name == "err" {
			return nil, errs.Errorf("sentinel")
		}
		return []string{"sym"}, nil
	})

	{
		got, err := ah.ConsumeFlag("foo", false)
		assert.NoError(t, err)
		assert.DeepEqual(t, got, []string{"bar"})
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "--bif", "--baf", "--zap=true", "arg", "--zap=false", "--zap", "--foo", "bing"})
	}

	{
		got, err := ah.ConsumeFlag("zap", false)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "--bif", "--baf", "--zap=true", "arg", "--zap=false", "--zap", "--foo", "bing"})
	}

	{
		got, err := ah.ConsumeFlag("zap", true)
		assert.NoError(t, err)
		assert.DeepEqual(t, got, []string{"true", "false", "true"})
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "--bif", "--baf", "arg", "--foo", "bing"})
	}

	{
		got, err := ah.ConsumeFlag("baf", false)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "--bif", "--baf", "arg", "--foo", "bing"})
	}

	{
		got, err := ah.ConsumeFlag("bif", false)
		assert.NoError(t, err)
		assert.DeepEqual(t, got, []string{"--baf"})
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "arg", "--foo", "bing"})
	}

	{
		got, err := ah.ConsumeFlag("not-exist", false)
		assert.NoError(t, err)
		assert.DeepEqual(t, got, []string{"sym"})
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "arg", "--foo", "bing"})
	}

	{
		got, err := ah.ConsumeFlag("err", false)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.DeepEqual(t, ah.PeekArgs(), []string{"baz", "arg", "--foo", "bing"})
	}

	{
		got, ok, err := ah.ConsumeArg("first")
		assert.NoError(t, err)
		assert.That(t, ok)
		assert.DeepEqual(t, got, "baz")
		assert.DeepEqual(t, ah.PeekArgs(), []string{"arg", "--foo", "bing"})
	}
}
