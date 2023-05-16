package clingy

import (
	"fmt"

	"github.com/zeebo/errs/v2"
)

type paramsPos struct {
	paramsTracker
	pm  *paramsMaker
	ah  *argsHandler
	opt bool // saw an optional argument
	rep bool // saw a repeated argument
}

func newParamsPositional(pm *paramsMaker, ah *argsHandler) *paramsPos {
	return &paramsPos{
		pm: pm,
		ah: ah,
	}
}

func (pp *paramsPos) Arg(name, desc string, options ...Option) (val interface{}) {
	p := pp.pm.newParam(name, desc, nil, options...)
	pp.include(p)

	// check for repeated/optional consistency
	if pp.opt && !(p.opt || p.rep) {
		panic(fmt.Sprintf("required argument after optional arguments: %q", name))
	}
	if pp.rep {
		panic(fmt.Sprintf("argument after repeated argument: %q", name))
	}
	pp.opt = pp.opt || p.opt
	pp.rep = pp.rep || p.rep

	if p.rep {
		val, p.err = pp.ah.ConsumeArgs()
		if p.err != nil {
			return p.zero()
		}
	} else {
		var ok bool
		val, ok, p.err = pp.ah.ConsumeArg()
		if p.err != nil {
			return p.zero()
		} else if !ok {
			if !p.opt {
				p.err = errs.Errorf("%s: required argument missing", name)
				return p.zero()
			}
			return p.zero()
		}
	}

	val, p.err = transformParam(p, val)
	return val
}
