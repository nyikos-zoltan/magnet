package gorm_v2

import (
	"reflect"

	"github.com/nyikos-zoltan/magnet"
	"gorm.io/gorm"
)

var gormType = reflect.TypeOf(&gorm.DB{})

type Transaction interface {
	Transaction(interface{}) error
}

type transaction struct {
	*magnet.Magnet
	db *gorm.DB
}

func (t *transaction) Transaction(fn interface{}) error {
	caller := t.NewCaller(fn, gormType)
	return t.db.Transaction(func(tx *gorm.DB) error {
		rv, err := caller.Call(tx)
		if err != nil {
			return nil
		}
		if err, ok := rv[0].Interface().(error); ok {
			return err
		}
		return nil
	})
}

func NewTransaction(parent *magnet.Magnet, db *gorm.DB) Transaction {
	return &transaction{parent.NewChild(), db}
}
