package magnet

import (
	"reflect"
)

func (f *Factory) AutoRef() {
	m := f.owner
	if f.provides.Kind() == reflect.Ptr {
		ftype := reflect.FuncOf(
			[]reflect.Type{f.provides},
			[]reflect.Type{f.provides.Elem()},
			false,
		)
		f := reflect.MakeFunc(ftype, func(in []reflect.Value) []reflect.Value {
			return []reflect.Value{in[0].Elem()}
		})
		m.Register(f.Interface())
	} else {
		ptrype := reflect.PtrTo(f.provides)
		ftype := reflect.FuncOf(
			[]reflect.Type{f.provides},
			[]reflect.Type{ptrype},
			false,
		)
		f := reflect.MakeFunc(ftype, func(in []reflect.Value) []reflect.Value {
			ptr := reflect.New(f.provides)
			ptr.Elem().Set(in[0])
			return []reflect.Value{ptr}
		})
		m.Register(f.Interface())
	}
}
