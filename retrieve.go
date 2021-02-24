package magnet

import (
	"fmt"
	"reflect"
)

func (m *Magnet) Retrieve(into interface{}) error {
	if reflect.TypeOf(into).Kind() != reflect.Ptr {
		panic("non-ptr type passed to Retrieve!")
	}
	reqValue := reflect.ValueOf(into).Elem()
	reqType := reflect.TypeOf(into).Elem()
	m.runHooks(reqType)
	m.detectCycles()
	if node := m.findNode(reqType); node != nil {
		val, err := node.Build(m)
		if err != nil {
			return err
		}
		reqValue.Set(val)
		return nil
	}
	panic(fmt.Errorf("unknown type %s", reqType))
}
