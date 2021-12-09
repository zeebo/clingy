package clingy

type paramsFlags struct {
	paramsTracker
	pm  *paramsMaker
	ah  *argsHandler
	err error
}

func newParamsFlags(ps *paramsMaker, ah *argsHandler) *paramsFlags {
	return &paramsFlags{
		pm: ps,
		ah: ah,
	}
}

func (pf *paramsFlags) Flag(name, desc string, def interface{}, options ...Option) (val interface{}) {
	p := pf.pm.newParam(name, desc, def, options...)
	if pf.err != nil {
		return p.zero()
	}
	pf.include(p)

	val, p.err = pf.getValue(p)
	if p.err != nil {
		return p.zero()
	} else if val == nil {
		if p.def == nil {
			return p.zero()
		}
		return p.def
	}

	val, p.err = transformParam(p, val)
	return val
}

func (pf *paramsFlags) getValue(p *param) (val interface{}, err error) {
	vals, err := pf.ah.ConsumeFlag(p.name, p.bstyle)
	if err != nil {
		return nil, err
	} else if vals == nil && p.short != 0 {
		vals, err = pf.ah.ConsumeFlag(string(p.short), p.bstyle)
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
