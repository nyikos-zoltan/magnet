package dbtx_test

import (
	"reflect"
	"testing"

	"github.com/nyikos-zoltan/magnet"
	"github.com/nyikos-zoltan/magnet/internal/dbtx"
	"github.com/nyikos-zoltan/magnet/transaction"
	"github.com/stretchr/testify/require"
)

type DB struct{ int }

func (d *DB) Transaction(cb func(*DB) error) error {
	d.int += 1
	return cb(&DB{d.int})
}

type Service struct {
	db *DB
}

type testerTx = func(func(transaction.Tx, Service) error) error

func Test_Dbtx(t *testing.T) {

	c := dbtx.DBTx{
		DBType: reflect.TypeOf((*DB)(nil)),
		Callback: func(c *magnet.Caller, dbI interface{}) error {
			db := dbI.(*DB)
			return db.Transaction(func(tx *DB) error {
				_, err := c.Call(tx, transaction.Tx{})
				return err
			})
		},
	}

	m := magnet.New()
	m.Register(func() *DB { return &DB{0} })
	m.Register(func(d *DB) Service { return Service{d} })

	m.RegisterTypeHook(c.SafeTxHook)

	var injS *Service
	call := m.NewCaller(func(t testerTx) error {
		return t(func(_ transaction.Tx, s Service) error {
			injS = &s
			return nil
		})
	})
	rv, err := call.Call()
	require.NoError(t, err)
	require.NoError(t, rv.Error(0))
	fdb := injS.db
	require.Equal(t, 1, injS.db.int)
	rv, err = call.Call()
	require.NoError(t, err)
	require.NoError(t, rv.Error(0))
	require.Equal(t, 2, injS.db.int)
	require.NotEqual(t, injS.db, fdb)
}
