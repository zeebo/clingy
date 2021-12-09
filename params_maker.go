package clingy

import "fmt"

type charSet [256 / 32]uint32

func (c *charSet) Set(x byte)      { c[x/32] |= 1 << (x % 32) }
func (c *charSet) Has(x byte) bool { return c[x/32]&(1<<(x%32)) != 0 }

type paramsMaker struct {
	set    map[string]*param
	shorts charSet
}

func newParamsMaker() *paramsMaker {
	return &paramsMaker{
		set: make(map[string]*param),
	}
}

func (ps *paramsMaker) newParam(name, desc string, def interface{}, options ...Option) *param {
	p := &param{name: name, def: def, desc: desc}
	for _, opt := range options {
		opt.do(&p.paramOpts)
	}
	if _, ok := ps.set[name]; ok {
		panic(fmt.Sprintf("parameter already defined with name: %q", name))
	} else if p.short != 0 && ps.shorts.Has(p.short) {
		panic(fmt.Sprintf("parameter already defined with short-name: %q", p.short))
	}
	var err error
	p.typ, err = checkFns(p.fns)
	if err != nil {
		panic(fmt.Sprintf("parameter has invalid transformation functions: %v", err))
	}
	ps.set[name] = p
	ps.shorts.Set(p.short)
	return p
}
