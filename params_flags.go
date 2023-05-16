package clingy

import (
	"fmt"

	"github.com/zeebo/errs/v2"
)

type paramsFlags struct {
	paramsTracker
	pm *paramsMaker
	ah *argsHandler
}

func newParamsFlags(ps *paramsMaker, ah *argsHandler) *paramsFlags {
	return &paramsFlags{
		pm: ps,
		ah: ah,
	}
}

func (pf *paramsFlags) Flag(name, desc string, def interface{}, options ...Option) (val interface{}) {
	p := pf.pm.newParam(name, desc, def, options...)
	pf.include(p)

	if p.opt && p.def == Required {
		panic(fmt.Sprintf("optional flag with Required default value: %q", name))
	}

	val, p.err = pf.getValue(p)
	if p.err != nil {
		return p.zero()
	} else if val == nil {
		if p.def == Required {
			p.err = errs.Errorf("%s: required flag missing", name)
			return p.zero()
		} else if p.def == nil {
			return p.zero()
		}
		return p.def
	}

	val, p.err = transformParam(p, val)
	return val
}

func (pf *paramsFlags) getValue(p *param) (val interface{}, err error) {
	vals, err := pf.ah.ConsumeFlag(p.name, p.bstyle, p.getenv)
	if err != nil {
		return nil, err
	} else if vals == nil && p.short != 0 {
		vals, err = pf.ah.ConsumeFlag(string(p.short), p.bstyle, p.getenv)
		if err != nil {
			return nil, err
		}
	}
	if p.rep {
		return vals, nil
	} else if len(vals) == 0 {
		return nil, nil
	} else {
		return vals[0], nil
	}
}
