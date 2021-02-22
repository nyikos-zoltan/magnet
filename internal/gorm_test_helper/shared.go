package gorm_test_helper

import (
	"database/sql"
	"database/sql/driver"
	"reflect"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nyikos-zoltan/magnet"
	"github.com/stretchr/testify/suite"
)

type GormSharedSuite struct {
	suite.Suite
	Magnet     *magnet.Magnet
	GormDB     interface{}
	GormDBType reflect.Type
	Mock       sqlmock.Sqlmock
}

type AnyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func (s *GormSharedSuite) MockDB(commit bool) {
	s.Mock.ExpectBegin()
	s.Mock.ExpectQuery("INSERT INTO \"?some_models\"?").WithArgs(AnyTime{}, AnyTime{}, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).FromCSVString("1"))
	if commit {
		s.Mock.ExpectCommit()
	} else {
		s.Mock.ExpectRollback()
	}
}

func (s *GormSharedSuite) SetupShared() {
	s.Magnet = magnet.New()
	var err error
	var db *sql.DB
	db, s.Mock, err = sqlmock.New()
	s.Magnet.Value(db)
	s.Require().NoError(err)
}

func (s *GormSharedSuite) RunTestTx(cb interface{}) error {
	tx := s.Magnet.NewCaller(cb)
	rv, err := tx.Call()
	s.Require().NoError(err)
	s.Require().Equal(1, rv.Len())
	return rv.Error(0)
}

func (s *GormSharedSuite) TearDownTest() {
	s.Require().NoError(s.Mock.ExpectationsWereMet())
}
