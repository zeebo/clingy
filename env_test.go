package clingy

import (
	"os"
	"testing"

	"github.com/zeebo/assert"
)

func TestEnv_FillDefaults(t *testing.T) {
	var env Environment
	env.fillDefaults()

	assert.That(t, env.Name != "")
	assert.NotNil(t, env.Args)
	assert.Equal(t, env.Stdin, os.Stdin)
	assert.Equal(t, env.Stdout, os.Stdout)
	assert.Equal(t, env.Stderr, os.Stderr)
	assert.NotNil(t, env.Getenv)
}
