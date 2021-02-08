package gorm_v1_test

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/c2fo/testify/require"
	"github.com/c2fo/testify/suite"
	"github.com/jinzhu/gorm"
	"github.com/nyikos-zoltan/magnet"
	"github.com/nyikos-zoltan/magnet/gorm_v1"
	"github.com/nyikos-zoltan/magnet/transaction"
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

type GormV1Suite struct {
	suite.Suite
	magnet *magnet.Magnet
	gormDB *gorm.DB
	mock   sqlmock.Sqlmock
}

func (s *GormV1Suite) SetupTest() {
	s.magnet = magnet.New()
	gorm_v1.Use(s.magnet)
	var err error
	var db *sql.DB
	db, s.mock, err = sqlmock.New()
	require.NoError(s.T(), err)

	s.mock.ExpectBegin()
	s.mock.ExpectQuery("INSERT INTO \"?some_models\"?").WithArgs(AnyTime{}, AnyTime{}, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).FromCSVString("1"))
	s.mock.ExpectCommit()

	s.gormDB, err = gorm.Open("postgres", db)
	s.magnet.Register(func() *gorm.DB { return s.gormDB })
	require.NoError(s.T(), err)
}

type TestTx = func(func(transaction.Transaction, *gorm.DB) error) error

var gormDBType = reflect.TypeOf((*gorm.DB)(nil))

func (s *GormV1Suite) TestOkCommit() {
	tx := s.magnet.NewCaller(func(tx TestTx) error {
		return tx(func(_ transaction.Transaction, txDB *gorm.DB) error {
			return txDB.Create(&SomeModel{}).Error
		})
	}, gormDBType)
	rv, err := tx.Call(s.gormDB)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, rv.Len())
	require.NoError(s.T(), rv.Error(0))
}

func (s *GormV1Suite) TestErrRollback() {
	tx := s.magnet.NewCaller(func(tx TestTx) error {
		return tx(func(_ transaction.Transaction, txDB *gorm.DB) error {
			txDB.Create(&SomeModel{})
			return errors.New("some error")
		})
	}, gormDBType)
	rv, err := tx.Call(s.gormDB)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, rv.Len())
	require.Error(s.T(), rv.Error(0))
}

func TestGormV1Suite(t *testing.T) {
	suite.Run(t, new(GormV1Suite))
}
