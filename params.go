package clingy

import (
	"reflect"
)

//
// paramater state information
//

type paramOpts struct {
	opt    bool
	rep    bool
	short  byte
	adv    bool
	hidden bool
	bstyle bool
	typ    string
	fns    []interface{}
}

type param struct {
	paramOpts
	name string
	def  interface{}
	desc string
	typ  reflect.Type
	err  error
}

func (p *param) zeroType() reflect.Type {
	typ := p.typ
	if p.opt && !p.rep {
		typ = reflect.PtrTo(typ)
	} else if p.rep {
		typ = reflect.SliceOf(typ)
	}
	return typ
}

func (p *param) zero() interface{} {
	return zero(p.zeroType())
}

func (p *param) flagType() string {
	if p.paramOpts.typ != "" {
		return p.paramOpts.typ
	}
	switch p.typ {
	case boolType:
		return ""
	case durationType:
		return "duration"
	default:
		return p.typ.Name()
	}
}

//
// wrapper to combine positional and flag parameters
//

type params struct {
	pp  *paramsPos
	pf  *paramsFlags
	pos bool
}

func newParams(pp *paramsPos, pf *paramsFlags) *params {
	return &params{
		pp: pp,
		pf: pf,
	}
}

func (p *params) Arg(name, desc string, options ...Option) (val interface{}) {
	p.pos = true
	return p.pp.Arg(name, desc, options...)
}

func (p *params) Flag(name, desc string, def interface{}, options ...Option) (val interface{}) {
	if p.pos {
		panic("must perform all Flag/Break calls before any Arg calls")
	}
	return p.pf.Flag(name, desc, def, options...)
}

func (p *params) Break() {
	if p.pos {
		panic("must perform all Flag/Break calls before any Arg calls")
	}
	p.pf.Break()
}
