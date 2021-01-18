package magnet

import (
	"reflect"
)

type Derived struct{}

var derivedType = reflect.TypeOf((*Derived)(nil)).Elem()

func (m *Magnet) registerDeriveds(types []reflect.Type) {
	for _, t := range types {
		if t.Kind() == reflect.Struct {
			sf, has := t.FieldByName("Derived")
			if has && sf.Type == derivedType {
				m.registerDerived(t)
			}
		}
	}
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

	m.providerMap[dsType] = &Node{
		requires: requires,
		provides: dsType,
		owner:    m,
		factory: reflect.ValueOf(func(params ...interface{}) reflect.Value {
			derivedValue := reflect.New(dsType).Elem()
			for i := 0; i < len(fieldIdx); i++ {
				derivedValue.Field(fieldIdx[i]).Set(reflect.ValueOf(params[i]))
			}
			return derivedValue
		}),
	}
}
