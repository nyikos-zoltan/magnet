package gorm_v2

import (
	"reflect"

	"github.com/nyikos-zoltan/magnet"
	"github.com/nyikos-zoltan/magnet/internal/dbtx"
	"github.com/nyikos-zoltan/magnet/transaction"
	"gorm.io/gorm"
)

var gormType = reflect.TypeOf(&gorm.DB{})

var gormv2DbTx = dbtx.DBTx{
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

func Use(m *magnet.Magnet) {
	m.RegisterTypeHook(gormv2DbTx.SafeTxHook)
}
