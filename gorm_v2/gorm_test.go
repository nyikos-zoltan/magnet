package gorm_v2_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/nyikos-zoltan/magnet/gorm_v2"
	"github.com/nyikos-zoltan/magnet/internal/gorm_test_helper"
	"github.com/nyikos-zoltan/magnet/transaction"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type SomeModel struct {
	gorm.Model
}

type GormV2Suite struct {
	gorm_test_helper.GormSharedSuite
}

func (s *GormV2Suite) SetupTest() {
	s.SetupShared()
	gorm_v2.Plugin(s.Magnet)

	s.Magnet.Register(func(db *sql.DB) *gorm.DB {
		gormDB, err := gorm.Open(postgres.New(postgres.Config{
			Conn: db,
		}), &gorm.Config{})
		s.Require().NoError(err)
		return gormDB
	})
}

type TestTx = func(func(transaction.Tx, *gorm.DB) error) error

func (s *GormV2Suite) TestOkCommit() {
	s.MockDB(true)
	err := s.RunTestTx(func(tx TestTx) error {
		return tx(func(_ transaction.Tx, txDB *gorm.DB) error {
			return txDB.Create(&SomeModel{}).Error
		})
	})
	s.Require().NoError(err)
}

func (s *GormV2Suite) TestErrRollback() {
	s.MockDB(false)
	err := s.RunTestTx(func(tx TestTx) error {
		return tx(func(_ transaction.Tx, txDB *gorm.DB) error {
			txDB.Create(&SomeModel{})
			return errors.New("some error")
		})
	})
	s.Require().Errorf(err, "some error")
}

func TestGormV2Suite(t *testing.T) {
	suite.Run(t, new(GormV2Suite))
}
