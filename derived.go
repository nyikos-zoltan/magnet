package magnet

import (
	"reflect"
)

type Derived struct{}

var derivedType = reflect.TypeOf((*Derived)(nil)).Elem()

func derivedTypeHook(h Hook, t reflect.Type) bool {
	if t.Kind() == reflect.Struct {
		sf, has := t.FieldByName("Derived")
		if has && sf.Type == derivedType {
			h.m.registerDerived(t)
			return true
		}
	}
	return false
}

func (m *Magnet) registerDerived(dsType reflect.Type) {
	var requires []reflect.Type
	var fieldIdx []int
	for i := 0; i < dsType.NumField(); i++ {
		f := dsType.Field(i)
		if f.Type == derivedType || f.PkgPath != "" {
			continue
		}
		requires = append(requires, f.Type)
		fieldIdx = append(fieldIdx, i)
	}

	factroryType := reflect.FuncOf(
		requires,
		[]reflect.Type{dsType},
		false,
	)

	factroryFn := reflect.MakeFunc(
		factroryType,
		func(in []reflect.Value) []reflect.Value {
			derivedValue := reflect.New(dsType).Elem()
			for i := 0; i < len(fieldIdx); i++ {
				derivedValue.Field(fieldIdx[i]).Set(in[i])
			}
			return []reflect.Value{derivedValue}
		},
	)

	m.providerMap[dsType] = &Node{
		requires: requires,
		provides: dsType,
		owner:    m,
		factory:  factroryFn,
	}
}
