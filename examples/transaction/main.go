package main

import (
	"fmt"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/nyikos-zoltan/magnet"
	"github.com/nyikos-zoltan/magnet/gorm_v1"
	"github.com/nyikos-zoltan/magnet/transaction"
)

func panic_if_error(err error) {
	if err != nil {
		panic(err)
	}
}

type Service struct {
	db *gorm.DB
}

type txType = func(func(transaction.Tx, *Service) error) error

func main() {
	db, mock, _ := sqlmock.New()

	m := magnet.New(gorm_v1.Plugin)
	m.Register(func() (*gorm.DB, error) {
		return gorm.Open("postgres", db)
	})
	m.Register(func(db *gorm.DB) *Service {
		return &Service{db}
	})
	mock.ExpectBegin()
	mock.ExpectCommit()

	var tx txType
	panic_if_error(m.Retrieve(&tx))

	err := tx(func(_ transaction.Tx, s *Service) error {
		// s.Something()
		return nil
	})
	panic_if_error(err)

	panic_if_error(mock.ExpectationsWereMet())
	fmt.Println("ok!")
}
