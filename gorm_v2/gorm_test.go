package gorm_v2_test

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nyikos-zoltan/magnet"
	"github.com/nyikos-zoltan/magnet/gorm_v2"
	"github.com/nyikos-zoltan/magnet/transaction"
	"github.com/nyikos-zoltan/magnet/tx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type SomeModel struct {
	gorm.Model
}

type AnyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

type GormV2Suite struct {
	suite.Suite
	magnet *magnet.Magnet
	gormDB *gorm.DB
	mock   sqlmock.Sqlmock
}

func (s *GormV2Suite) SetupTest() {
	s.magnet = magnet.New()
	gorm_v2.Use(s.magnet)
	var err error
	var db *sql.DB
	db, s.mock, err = sqlmock.New()
	require.NoError(s.T(), err)

	s.mock.ExpectBegin()
	s.mock.ExpectQuery("INSERT INTO \"?some_models\"?").WithArgs(AnyTime{}, AnyTime{}, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).FromCSVString("1"))
	s.mock.ExpectCommit()

	s.gormDB, err = gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	s.magnet.Register(func() *gorm.DB { return s.gormDB })
	require.NoError(s.T(), err)
}

type TestTx = func(func(tx.Tx, *gorm.DB) error) error

var gormDBType = reflect.TypeOf((*gorm.DB)(nil))

func (s *GormV2Suite) TestOkCommit() {
	tx := s.magnet.NewCaller(func(tx TestTx) error {
		return tx(func(_ transaction.Tx, txDB *gorm.DB) error {
			return txDB.Create(&SomeModel{}).Error
		})
	}, gormDBType)
	rv, err := tx.Call(s.gormDB)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, rv.Len())
	require.NoError(s.T(), rv.Error(0))
}

func (s *GormV2Suite) TestErrRollback() {
	tx := s.magnet.NewCaller(func(tx TestTx) error {
		return tx(func(_ tx.Tx, txDB *gorm.DB) error {
			txDB.Create(&SomeModel{})
			return errors.New("some error")
		})
	}, gormDBType)
	rv, err := tx.Call(s.gormDB)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, rv.Len())
	require.Error(s.T(), rv.Error(0))
}

func TestGormV2Suite(t *testing.T) {
	suite.Run(t, new(GormV2Suite))
}
