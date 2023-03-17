package db

import (
	"errors"
	"github.com/jackc/pgx/v4"
)

var dbase *DBase = nil

func Init(dbUri string) error {
	if dbUri == "" {
		return errors.New("empty DB url")
	}
	if dbase != nil && dbase.connection != nil {
		dbase.connection.Close()
	}

	var err error
	dbase, err = Create(dbUri)
	if err != nil {
		return err
	}
	if dbase == nil {
		return errors.New("failed to initialize DB")
	}
	return nil
}

func QueryRow(query string, args ...interface{}) (pgx.Row, error) {
	if dbase == nil {
		return nil, errors.New("database connection is not created")
	}
	return dbase.QueryRow(query, args...)
}

func QueryRows(query string, args ...interface{}) (pgx.Rows, error) {
	if dbase == nil {
		return nil, errors.New("database connection is not created")
	}
	return dbase.QueryRows(query, args...)
}

func IncrementExec(query string, args ...interface{}) (int, error) {
	if dbase == nil {
		return 0, errors.New("database connection is not created")
	}
	return dbase.IncrementExec(query, args...)
}

func Exec(query string, args ...interface{}) (int, error) {
	if dbase == nil {
		return 0, errors.New("database connection is not created")
	}
	return dbase.Exec(query, args...)
}

func BulkInsert(table string, columns []string, rows [][]interface{}) (int, error) {
	if dbase == nil {
		return 0, errors.New("database connection is not created")
	}
	return dbase.BulkInsert(table, columns, rows)
}
