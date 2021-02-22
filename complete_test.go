package magnet_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/nyikos-zoltan/magnet"
	"github.com/nyikos-zoltan/magnet/internal/dbtx"
	"github.com/nyikos-zoltan/magnet/transaction"
	"github.com/stretchr/testify/require"
)

type DB struct {
	isTx   bool
	isDone bool
}

func (d *DB) Transaction(cb func(*DB) error) error {
	db := DB{isTx: true, isDone: false}
	err := cb(&db)
	db.isDone = true
	return err
}

func (d *DB) Exec() error {
	if d.isDone {
		return errors.New("already done!")
	} else {
		return nil
	}
}

type Repo interface {
	Query() error
}
type repo struct {
	db *DB
}

type Service interface {
	Do() error
}

type service struct {
	r Repo
}

func (s *service) Do() error {
	return s.r.Query()
}

func (s *repo) Query() error {
	return s.db.Exec()
}

var dbTXDef = dbtx.DBTx{
	DBType: reflect.TypeOf((*DB)(nil)),
	Callback: func(c *magnet.Caller, dbI interface{}) error {
		db := dbI.(*DB)
		return db.Transaction(func(tx *DB) error {
			rv, err := c.Call(tx, transaction.Tx{})
			if err != nil {
				return err
			}
			return rv.Error(0)
		})
	},
}

type testTx = func(func(transaction.Tx, testTxDeps) error) error

type HandlerDeps struct {
	magnet.Derived
	Tx testTx
}

type testTxDeps struct {
	magnet.Derived
	S Service
}

func Test_Complete(t *testing.T) {
	m := magnet.New()
	m.RegisterTypeHook(dbTXDef.SafeTxHook)
	m.Register(func() *DB { return &DB{} })
	m.Register(func(d *DB) Repo { return &repo{d} })
	m.Register(func(r Repo) Service { return &service{r} })
	ctx := echo.New().NewContext(nil, nil)
	h2 := m.EchoHandler(func(e echo.Context, d HandlerDeps) error {
		return d.Tx(func(_ transaction.Tx, t testTxDeps) error {
			return t.S.Do()
		})
	})
	require.NoError(t, h2(ctx))
	require.NoError(t, h2(ctx))
}
