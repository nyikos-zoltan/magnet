package gorm_v1

import (
	"reflect"

	"github.com/jinzhu/gorm"
	"github.com/nyikos-zoltan/magnet"
)

var gormType = reflect.TypeOf(&gorm.DB{})
var errType = reflect.TypeOf((*error)(nil)).Elem()
var magnetType = reflect.TypeOf(&magnet.Magnet{})

type Transaction struct{}

var safeTxType = reflect.TypeOf(Transaction{})

func safeTxHook(h magnet.Hook, txType reflect.Type) bool {
	if txType.Kind() == reflect.Func {
		if txType.NumIn() == 1 && txType.NumOut() == 1 {
			cbType := txType.In(0)
			if cbType.Kind() == reflect.Func && txType.Out(0) == errType && cbType.NumOut() == 1 && cbType.Out(0) == errType {
				var requires []reflect.Type
				has := false

				for i := 0; i < cbType.NumIn(); i++ {
					param := cbType.In(i)
					if param == safeTxType {
						has = true
					} else {
						requires = append(requires, param)
					}
				}

				if !has {
					return false
				}

				h.ValidateDeps(requires)
				factoryParams := []reflect.Type{
					magnetType,
					gormType,
				}
				factoryType := reflect.FuncOf(
					factoryParams,
					[]reflect.Type{txType},
					false,
				)
				factory := reflect.MakeFunc(
					factoryType,
					func(in []reflect.Value) []reflect.Value {
						m := in[0].Interface().(*magnet.Magnet)
						db := in[1].Interface().(*gorm.DB)

						rv := reflect.MakeFunc(
							txType,
							func(in []reflect.Value) []reflect.Value {
								fn := in[0].Interface()
								caller := m.NewCaller(fn, gormType, safeTxType)
								rv := db.Transaction(func(tx *gorm.DB) error {
									rv, err := caller.Call(tx, Transaction{})
									if err != nil {
										return err
									}
									if err, ok := rv[0].Interface().(error); ok {
										return err
									}
									return nil
								})
								if rv == nil {
									return []reflect.Value{reflect.Zero(errType)}
								} else {
									return []reflect.Value{reflect.ValueOf(rv)}
								}
							},
						)
						return []reflect.Value{rv}
					},
				)

				h.RegisterNewType(
					factoryParams,
					txType,
					factory.Interface(),
				)

				return true
			}
		}
	}
	return false
}

func Use(m *magnet.Magnet) {
	m.RegisterTypeHook(safeTxHook)
}
