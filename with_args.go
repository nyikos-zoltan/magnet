package magnet

import "reflect"

func WithArgs(fn interface{}, args ...interface{}) interface{} {
	fval := reflect.ValueOf(fn)
	ftype := fval.Type()
	var argvals []reflect.Value
	for _, arg := range args {
		argvals = append(argvals, reflect.ValueOf(arg))
	}

	var rtypes []reflect.Type
	for i := 0; i < ftype.NumOut(); i++ {
		rtypes = append(rtypes, ftype.Out(i))
	}
	nftype := reflect.FuncOf([]reflect.Type{}, rtypes, false)
	nfval := reflect.MakeFunc(
		nftype,
		func([]reflect.Value) []reflect.Value {
			return fval.Call(argvals)
		},
	)
	return nfval.Interface()
}
