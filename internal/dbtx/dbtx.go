// Package dbtx is shared code between gorm_v1 and gorm_v2, possibly supporting other db drivers in the future
package dbtx

import (
	"reflect"

	"github.com/nyikos-zoltan/magnet"
	"github.com/nyikos-zoltan/magnet/transaction"
)

type DBTx struct {
	DBType   reflect.Type
	Callback func(*magnet.Caller, interface{}) error
}

var safeTxType = reflect.TypeOf(transaction.Tx{})
var errType = reflect.TypeOf((*error)(nil)).Elem()
var magnetType = reflect.TypeOf(&magnet.Magnet{})

func (txDef *DBTx) SafeTxHook(h magnet.Hook, txType reflect.Type) bool {
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
					txDef.DBType,
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
						dbI := in[1].Interface()

						rv := reflect.MakeFunc(
							txType,
							func(in []reflect.Value) []reflect.Value {
								fn := in[0].Interface()
								caller := m.NewCaller(fn, txDef.DBType, safeTxType)

								rv := txDef.Callback(caller, dbI)

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
