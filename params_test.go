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
		ah    = newArgsHandler([]string{"foo", "--int", "100", "true", "10", "20", "30"}, nil, nil)
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

func TestShortFlagOverridesFallback(t *testing.T) {
	for _, source := range []string{"environment", "dynamic"} {
		t.Run(source, func(t *testing.T) {
			fallback := func(string) string { t.Fatal("unexpected environment lookup"); return "env" }
			dynamic := func(string) ([]string, error) { t.Fatal("unexpected dynamic lookup"); return []string{"dynamic"}, nil }
			env := "NAME"
			if source == "dynamic" {
				env = ""
			}
			ah := newArgsHandler([]string{"-n", "cli"}, dynamic, fallback)
			pf := newParamsFlags(newParamsMaker(), ah)
			assert.Equal(t, pf.Flag("name", "", "default", Short('n'), Getenv(env)), "cli")
			rest, err := ah.ConsumeArgs()
			assert.NoError(t, err)
			assert.Equal(t, len(rest), 0)
		})
	}
}

func TestRepeatedFlagMixedAliases(t *testing.T) {
	for _, args := range [][]string{
		{"--name", "a", "-n", "b", "--name=c", "-n=d"},
		{"-n=a", "--name=b", "-n", "c", "--name", "d"},
	} {
		ah := newArgsHandler(args, nil, nil)
		pf := newParamsFlags(newParamsMaker(), ah)
		assert.DeepEqual(t, pf.Flag("name", "", nil, Short('n'), Repeated), []string{"a", "b", "c", "d"})
		rest, err := ah.ConsumeArgs()
		assert.NoError(t, err)
		assert.Equal(t, len(rest), 0)
	}
}

func TestFlagAliasNameCollision(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		pf := newParamsFlags(newParamsMaker(), newArgsHandler(nil, nil, nil))
		first := func() { pf.Flag("name", "", "", Short('n')) }
		second := func() { pf.Flag("n", "", "") }
		if reverse {
			first, second = second, first
		}
		first()
		func() {
			defer func() {
				if recover() == nil {
					t.Error("accepted conflicting name and alias")
				}
			}()
			second()
		}()
	}
	// Giving a flag its own single-letter name as an alias is unambiguous.
	pf := newParamsFlags(newParamsMaker(), newArgsHandler([]string{"-n", "value"}, nil, nil))
	assert.Equal(t, pf.Flag("n", "", "", Short('n')), "value")
}
