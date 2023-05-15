package clingy_test

import (
	"testing"

	"github.com/zeebo/clingy"
)

func TestSummary(t *testing.T) {
	cmds := func(cmds clingy.Commands) {
		cmds.New("cmd1", "cmd1 short", printCommand("cmd1"))
		cmds.New("cmd2", "cmd2 short", printCommand("cmd2"))
		cmds.Group("group1", "group1 short", func() {
			cmds.New("sub1", "sub1 short", printCommand("group1 sub1"))
			cmds.Group("group2", "group2 short", func() {
				cmds.New("sub2", "sub2 short", printCommand("group1 group2 sub2"))
			})
			cmds.New("sub3", "sub3 short", printCommand("group1 sub3"))
		})
		cmds.New("cmd3", "cmd3 short", printCommand("cmd3"))
		cmds.New("cmd4", "cmd4 short", printCommand("cmd4"))
	}

	{
		result := Capture(Env("cmd", nil, "--summary"), cmds)
		result.AssertValid(t)
		result.AssertStdout(t, `
			Available commands:
			    cmd cmd1                  cmd1 short
			    cmd cmd2                  cmd2 short
			    cmd group1 sub1           sub1 short
			    cmd group1 group2 sub2    sub2 short
			    cmd group1 sub3           sub3 short
			    cmd cmd3                  cmd3 short
			    cmd cmd4                  cmd4 short
		`)
	}

	{
		result := Capture(Env("cmd", nil, "group1", "--summary"), cmds)
		result.AssertValid(t)
		result.AssertStdout(t, `
			Available commands:
			    cmd group1 sub1           sub1 short
			    cmd group1 group2 sub2    sub2 short
			    cmd group1 sub3           sub3 short
		`)
	}

	{
		result := Capture(Env("cmd", nil, "group1", "group2", "--summary"), cmds)
		result.AssertValid(t)
		result.AssertStdout(t, `
			Available commands:
			    cmd group1 group2 sub2    sub2 short
		`)
	}
}
