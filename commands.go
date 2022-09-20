package clingy

import (
	"strings"
)

type cmdDesc struct {
	name    string
	short   string
	long    string
	cmd     Command
	subcmds []cmdDesc
}

type commands struct {
	Flags
	cur []cmdDesc
}

func newCommands(flags Flags) *commands {
	return &commands{Flags: flags}
}

func (cmds *commands) collect(fn func()) (out []cmdDesc) {
	out, cmds.cur = cmds.cur, out
	if fn != nil {
		fn()
	}
	out, cmds.cur = cmds.cur, out
	return out
}

func (cmds *commands) New(name, desc string, cmd Command) {
	short, long := parseDesc(desc)
	cmds.cur = append(cmds.cur, cmdDesc{
		name:  name,
		short: short,
		long:  long,
		cmd:   cmd,
	})
}

func (cmds *commands) Group(name, desc string, children func()) {
	cmds.cur = append(cmds.cur, cmdDesc{
		name:    name,
		short:   desc,
		subcmds: cmds.collect(children),
	})
}

func collectDescs(flags Flags, fn func(Commands)) []cmdDesc {
	cmds := newCommands(flags)
	if fn == nil {
		fn = func(Commands) {}
	}
	return cmds.collect(func() { fn(cmds) })
}

func parseDesc(desc string) (short, long string) {
	desc = strings.TrimSpace(desc)
	idx := strings.IndexByte(desc, '\n')
	if idx == -1 {
		return desc, ""
	}
	short, desc = strings.TrimSpace(desc[:idx]), desc[idx+1:]

	lines := strings.Split(desc, "\n")
	minWhite := -1
	for i, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			lines[i] = ""
			continue
		}
		white := 0
		for ; white < len(line); white++ {
			if line[white] != ' ' && line[white] != '\t' {
				break
			}
		}
		if white < minWhite || minWhite == -1 {
			minWhite = white
		}
	}
	if minWhite > 0 {
		for i, line := range lines {
			if line != "" {
				lines[i] = line[minWhite:]
			}
		}
	}
	return short, strings.TrimSpace(strings.Join(lines, "\n"))
}
