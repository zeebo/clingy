package clingy

import (
	"testing"

	"github.com/zeebo/assert"
)

func TestCollectDescs(t *testing.T) {
	assert.DeepEqual(t, collectDescs(nil, func(c Commands, flags Flags) {
		c.New("foo0", "foo0", nil)
		c.Group("bar", "bar", func() {
			c.New("bar0", "bar0", nil)
			c.New("bar1", "bar1", nil)
			c.Group("baz", "baz", func() {
				c.New("baz0", "baz0", nil)
			})
			c.New("bar2", "bar2", nil)
		})
		c.New("foo1", "foo1", nil)
	}), []cmdDesc{
		{"foo0", "foo0", "", nil, nil},
		{"bar", "bar", "", nil, []cmdDesc{
			{"bar0", "bar0", "", nil, nil},
			{"bar1", "bar1", "", nil, nil},
			{"baz", "baz", "", nil, []cmdDesc{
				{"baz0", "baz0", "", nil, nil},
			}},
			{"bar2", "bar2", "", nil, nil},
		}},
		{"foo1", "foo1", "", nil, nil},
	})
}
