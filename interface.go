package magnet

import (
	"reflect"
)

func InterfacePlugin(m *Magnet) {
	m.RegisterTypeHook(func(h Hook, t reflect.Type) bool {
		if t.Kind() == reflect.Interface {
			node := h.m.findNodeByPred(func(ntype reflect.Type) bool {
				return t != ntype && ntype.Implements(t)
			})
			if node == nil {
				return false
			}
			ftype := reflect.FuncOf(
				[]reflect.Type{node.provides},
				[]reflect.Type{t},
				false,
			)
			factory := reflect.MakeFunc(ftype, func(in []reflect.Value) []reflect.Value {
				return in
			})
			m.Register(factory.Interface())
		}
		return false
	})
}
