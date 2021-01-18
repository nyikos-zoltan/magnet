package gorm_v2_test

import (
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/c2fo/testify/require"
	"github.com/nyikos-zoltan/magnet"
	"github.com/nyikos-zoltan/magnet/gorm_v2"
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

func Test_GormV1_Transaction(t *testing.T) {
	t.Run("ok - commit", func(t *testing.T) {
		m := magnet.New()
		db, mock, err := sqlmock.New()
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO \"?some_models\"?").WithArgs(AnyTime{}, AnyTime{}, nil).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).FromCSVString("1"))
		mock.ExpectCommit()
		require.NoError(t, err)
		gormDB, err := gorm.Open(postgres.New(postgres.Config{
			Conn: db,
		}), &gorm.Config{})
		require.NoError(t, err)
		tx := gorm_v2.NewTransaction(m, gormDB)

		require.NoError(t, tx.Transaction(func(txDB *gorm.DB) error {
			return txDB.Create(&SomeModel{}).Error
		}))
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
	t.Run("err - rollback", func(t *testing.T) {
		m := magnet.New()
		db, mock, err := sqlmock.New()
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO \"?some_models\"?").WithArgs(AnyTime{}, AnyTime{}, nil).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).FromCSVString("1"))
		mock.ExpectRollback()
		require.NoError(t, err)
		gormDB, err := gorm.Open(postgres.New(postgres.Config{
			Conn: db,
		}), &gorm.Config{})
		require.NoError(t, err)
		tx := gorm_v2.NewTransaction(m, gormDB)

		err = tx.Transaction(func(txDB *gorm.DB) error {
			txDB.Create(&SomeModel{})
			return errors.New("some error")
		})
		require.EqualError(t, err, "some error")
		if err = mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}
