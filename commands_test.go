package clingy

import (
	"testing"

	"github.com/zeebo/assert"
)

func TestCollectDescs(t *testing.T) {
	assert.DeepEqual(t, collectDescs(nil, func(cmds Commands) {
		cmds.New("foo0", "foo0", nil)
		cmds.Group("bar", "bar", func() {
			cmds.New("bar0", "bar0", nil)
			cmds.New("bar1", "bar1", nil)
			cmds.Group("baz", "baz", func() {
				cmds.New("baz0", "baz0", nil)
			})
			cmds.New("bar2", "bar2", nil)
		})
		cmds.New("foo1", "foo1", nil)
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
