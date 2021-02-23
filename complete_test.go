package magnet_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/go-resty/resty/v2"
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
	DBType: reflect.TypeOf(&DB{}),
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
	P  HandlerParam
	E  echo.Context
}

type testTxDeps struct {
	magnet.Derived
	S Service
}

type HandlerParam struct {
	Data string `json:"data"`
}

type HandlerResult struct {
	Data string `json:"data"`
}

func Test_Complete(t *testing.T) {
	m := magnet.New()
	m.RegisterTypeHook(dbTXDef.SafeTxHook)
	m.RegisterMany(
		func() *DB { return &DB{} },
		func(d *DB) Repo { return &repo{d} },
		func(r Repo) Service { return &service{r} },
		func(e echo.Context) (p HandlerParam, err error) {
			err = e.Bind(&p)
			return
		},
	)
	e := echo.New()
	e.HideBanner, e.HidePort = true, true
	e.POST("/", m.EchoHandler(func(d HandlerDeps) error {
		err := d.Tx(func(_ transaction.Tx, t testTxDeps) error {
			return t.S.Do()
		})
		if err != nil {
			return err
		}
		return d.E.JSON(200, HandlerResult(d.P))
	}))

	go func() {
		require.Errorf(t, e.Start("127.0.0.1:19999"), "http: Server closed")
	}()
	c := resty.New()
	resp, err := c.R().SetBody(HandlerParam{Data: "x"}).SetResult(HandlerResult{}).Post("http://127.0.0.1:19999")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode())
	require.Equal(t, &HandlerResult{Data: "x"}, resp.Result())

	resp, err = c.R().SetBody(HandlerParam{Data: "y"}).SetResult(HandlerResult{}).Post("http://127.0.0.1:19999")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode())
	require.Equal(t, &HandlerResult{Data: "y"}, resp.Result())

	require.NoError(t, e.Shutdown(context.Background()))
}
