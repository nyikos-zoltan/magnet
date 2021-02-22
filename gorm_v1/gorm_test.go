package gorm_v1_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/nyikos-zoltan/magnet/gorm_v1"
	"github.com/nyikos-zoltan/magnet/internal/gorm_test_helper"
	"github.com/nyikos-zoltan/magnet/transaction"
	"github.com/stretchr/testify/suite"
)

type SomeModel struct {
	gorm.Model
}

type GormV1Suite struct {
	gorm_test_helper.GormSharedSuite
}

func (s *GormV1Suite) SetupTest() {
	s.SetupShared()
	gorm_v1.Plugin(s.Magnet)
	s.Magnet.Register(func(db *sql.DB) *gorm.DB {
		gorm, err := gorm.Open("postgres", db)
		s.Require().NoError(err)
		return gorm
	})
}

type TestTx = func(func(transaction.Tx, *gorm.DB) error) error

func (s *GormV1Suite) TestOkCommit() {
	s.MockDB(true)
	err := s.RunTestTx(func(tx TestTx) error {
		return tx(func(_ transaction.Tx, txDB *gorm.DB) error {
			return txDB.Create(&SomeModel{}).Error
		})
	})
	s.Require().NoError(err)
}

func (s *GormV1Suite) TestErrRollback() {
	s.MockDB(false)
	err := s.RunTestTx(func(tx TestTx) error {
		return tx(func(_ transaction.Tx, txDB *gorm.DB) error {
			txDB.Create(&SomeModel{})
			return errors.New("some error")
		})
	})
	s.Require().Errorf(err, "some error")
}

func TestGormV1Suite(t *testing.T) {
	suite.Run(t, new(GormV1Suite))
}
