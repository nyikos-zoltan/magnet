package testmagnet

import (
	"reflect"

	"github.com/nyikos-zoltan/magnet"
	"github.com/nyikos-zoltan/magnet/internal/dbtx"
	"github.com/nyikos-zoltan/magnet/transaction"
)

type fake struct{}

var dbType = reflect.TypeOf(fake{})

var testDbTx = dbtx.DBTx{
	DBType: dbType,
	Callback: func(c *magnet.Caller, dbI interface{}) error {
		rv, err := c.Call(fake{}, transaction.Tx{})
		if err != nil {
			return err
		}
		return rv.Error(0)
	},
}

// TxPlugin configures magnet to provide transactions in testing, by simpling providing the same objects into the transaction
func TxPlugin(m *magnet.Magnet) {
	m.Value(fake{})
	m.RegisterTypeHook(testDbTx.SafeTxHook)
}
