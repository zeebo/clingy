package clingy

import (
	"strconv"
	"testing"

	"github.com/zeebo/assert"
)

func TestParams(t *testing.T) {
	var (
		parseBool = Transform(strconv.ParseBool)
		parseInt  = Transform(strconv.Atoi)

		pos   = newParamsPositional([]string{"foo", "true", "10", "20", "30"})
		flags = newParamsFlags(map[string][]string{"int": {"100"}}, nil)
	)

	tr := true

	assert.DeepEqual(t, "foo", pos.Arg("string", "").(string))
	assert.DeepEqual(t, &tr, pos.Arg("bool", "", Optional, parseBool).(*bool))
	assert.DeepEqual(t, []int{10, 20, 30}, pos.Arg("repInt", "", Repeated, parseInt).([]int))
	assert.DeepEqual(t, 100, flags.Flag("int", "", 5, parseInt).(int))
	assert.DeepEqual(t, 5, flags.Flag("def", "", 5, parseInt).(int))
}
