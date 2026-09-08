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

func (pf *paramsFlags) Flag(name, desc string, def any, options ...Option) (val any) {
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

func (pf *paramsFlags) getValue(p *param) (val any, err error) {
	names := []string{p.name}
	if p.short != 0 {
		names = append(names, string(p.short))
	}
	vals, err := pf.ah.consumeFlags(names, p.bstyle)
	if err != nil {
		return nil, err
	}
	if vals == nil {
		vals, err = pf.ah.flagDefault(p.name, p.getenv)
		if err != nil {
			return nil, err
		}
	}
	if p.rep {
		// vals is a slice of strings and the following are not equal:
		//    var x any = []string(nil)
		//    var x any = nil
		if vals == nil {
			return nil, nil
		}
		return vals, nil
	} else if len(vals) == 0 {
		return nil, nil
	} else {
		return vals[0], nil
	}
}
