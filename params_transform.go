package clingy

import (
	"reflect"
	"time"

	"github.com/zeebo/errs/v2"
)

func transformParam(arg *param, val interface{}) (_ interface{}, err error) {
	call := callOne
	if arg.rep {
		call = callMany
	}

	rval := reflect.ValueOf(val)
	for _, fn := range arg.fns {
		rval, err = call(rval, reflect.ValueOf(fn))
		if err != nil {
			return arg.zero(), err
		}
	}

	if arg.opt && !arg.rep {
		rval = ptrTo(rval)
	}
	return rval.Interface(), nil
}

func callMany(rval, rfn reflect.Value) (reflect.Value, error) {
	if rval.IsNil() {
		return reflect.Zero(reflect.SliceOf(rfn.Type().Out(0))), nil
	}
	out := reflect.MakeSlice(reflect.SliceOf(rfn.Type().Out(0)), rval.Len(), rval.Len())
	for i := 0; i < rval.Len(); i++ {
		result, err := callOne(rval.Index(i), rfn)
		if err != nil {
			return reflect.Value{}, err
		}
		out.Index(i).Set(result)
	}
	return out, nil
}

func callOne(rval, rfn reflect.Value) (reflect.Value, error) {
	results := rfn.Call([]reflect.Value{rval})
	if err, ok := results[1].Interface().(error); ok && err != nil {
		return reflect.Value{}, err
	}
	return results[0], nil
}

var (
	stringType   = reflect.TypeOf("")
	boolType     = reflect.TypeOf(false)
	durationType = reflect.TypeOf(time.Duration(0))
	errorType    = reflect.TypeOf((*error)(nil)).Elem()
)

func guessType(fns []interface{}) reflect.Type {
	typ := stringType
	if len(fns) > 0 {
		ftyp := reflect.TypeOf(fns[len(fns)-1])
		if ftyp.Kind() == reflect.Func && ftyp.NumOut() > 0 {
			typ = ftyp.Out(0)
		}
	}
	return typ
}

func zero(typ reflect.Type) interface{} {
	return reflect.Zero(typ).Interface()
}

func ptrTo(x reflect.Value) reflect.Value {
	y := reflect.New(x.Type())
	y.Elem().Set(x)
	return y
}

func checkFns(fns []interface{}) (reflect.Type, error) {
	typ := stringType
	for _, fn := range fns {
		ftyp := reflect.TypeOf(fn)
		switch {
		case false,
			ftyp.Kind() != reflect.Func,
			ftyp.NumIn() != 1,
			ftyp.NumOut() != 2,
			ftyp.In(0) != typ,
			ftyp.Out(1) != errorType:
			return guessType(fns), errs.Errorf("transform: %v cannot be applied to %v", ftyp, typ)
		}
		typ = ftyp.Out(0)
	}
	return typ, nil
}
