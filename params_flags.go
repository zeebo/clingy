package clingy

type paramsFlags struct {
	ps      *paramsShared
	flags   map[string][]string
	dynamic func(string) ([]string, error)
	dynerr  error
}

func newParamsFlags(flags map[string][]string, dynamic func(string) ([]string, error)) *paramsFlags {
	return &paramsFlags{ps: newParamsShared(), flags: flags, dynamic: dynamic}
}

func (ps *paramsFlags) count() int             { return ps.ps.count() }
func (ps *paramsFlags) params(cb func(*param)) { ps.ps.iter(cb) }
func (ps *paramsFlags) hasErrors() bool        { return ps.ps.hasErrors() }

func (ps *paramsFlags) Flag(name, desc string, def interface{}, options ...Option) (val interface{}) {
	p := ps.ps.newParam(name, desc, def, options...)
	if ps.dynerr != nil {
		return p.zero()
	}

	val, p.err = ps.getValue(p)
	if p.err != nil {
		return p.zero()
	} else if val == nil {
		return p.def
	}

	val, p.err = transformParam(p, val)
	return val
}

func (ps *paramsFlags) getValue(p *param) (val interface{}, err error) {
	vals, ok := ps.flags[p.name]
	if !ok && p.short != 0 {
		vals, ok = ps.flags[string(p.short)]
	}
	if !ok && ps.dynamic != nil {
		vals, ps.dynerr = ps.dynamic(p.name)
		if err != nil {
			return nil, ps.dynerr
		}
	}
	if p.rep {
		return vals, nil
	} else if len(vals) == 0 {
		if p.def != nil {
			return nil, nil
		}
		return "", nil
	} else {
		return vals[0], nil
	}
}
