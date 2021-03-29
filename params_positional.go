package clingy

import (
	"fmt"

	"github.com/zeebo/errs/v2"
)

type paramsPos struct {
	ps  *params
	pos []string
	opt bool // saw an optional argument
	rep bool // saw a repeated argument
}

func newParamsPositional(pos []string) *paramsPos {
	return &paramsPos{ps: newParams(), pos: pos}
}

func (ps *paramsPos) params(cb func(*param)) { ps.ps.iter(cb) }
func (ps *paramsPos) hasErrors() bool        { return ps.ps.hasErrors() }

func (ps *paramsPos) New(name, desc string, options ...Option) (val interface{}) {
	p := ps.ps.newParam(name, desc, nil, options...)
	if p.err != nil {
		return p.zero()
	}

	// check for repeated/optional consistency
	if ps.opt && !(p.opt || p.rep) {
		panic(fmt.Sprintf("required argument after optional arguments: %q", name))
	}
	if ps.rep {
		panic(fmt.Sprintf("argument after repeated argument: %q", name))
	}
	ps.opt = ps.opt || p.opt
	ps.rep = ps.rep || p.rep

	if p.rep {
		val, ps.pos = ps.pos, nil
	} else if len(ps.pos) == 0 {
		if !p.opt {
			p.err = errs.Errorf("%s: required argument missing", name)
			return p.zero()
		}
		return p.zero()
	} else {
		val, ps.pos = ps.pos[0], ps.pos[1:]
	}
	val, p.err = transformParam(p, val)
	return val
}
