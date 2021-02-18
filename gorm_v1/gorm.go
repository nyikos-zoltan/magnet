package gorm_v1

import (
	"reflect"

	"github.com/jinzhu/gorm"
	"github.com/nyikos-zoltan/magnet"
	"github.com/nyikos-zoltan/magnet/internal/dbtx"
	"github.com/nyikos-zoltan/magnet/transaction"
)

var gormType = reflect.TypeOf(&gorm.DB{})

var gormv1DbTx = dbtx.DBTx{
	DBType: gormType,
	Callback: func(c *magnet.Caller, dbI interface{}) error {
		db := dbI.(*gorm.DB)
		return db.Transaction(func(tx *gorm.DB) error {
			rv, err := c.Call(tx, transaction.Tx{})
			if err != nil {
				return err
			}
			return rv.Error(0)
		})
	},
}

func Plugin(m *magnet.Magnet) {
	m.RegisterTypeHook(gormv1DbTx.SafeTxHook)
}
