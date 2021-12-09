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

		pm    = newParamsMaker()
		ah    = newArgsHandler([]string{"foo", "--int", "100", "true", "10", "20", "30"}, nil)
		pos   = newParamsPositional(pm, ah)
		flags = newParamsFlags(pm, ah)
	)

	tr := true

	assert.DeepEqual(t, 100, flags.Flag("int", "", 5, parseInt).(int))
	assert.DeepEqual(t, 5, flags.Flag("def", "", 5, parseInt).(int))
	assert.DeepEqual(t, "foo", pos.Arg("string", "").(string))
	assert.DeepEqual(t, &tr, pos.Arg("bool", "", Optional, Boolean, parseBool).(*bool))
	assert.DeepEqual(t, []int{10, 20, 30}, pos.Arg("repInt", "", Repeated, parseInt).([]int))
}
