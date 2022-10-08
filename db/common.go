package db

import (
	"context"
	"errors"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
	"math"
	"time"
)

var dbConnection *pgxpool.Pool = nil

func Init(dbUri string) error {

	var err error

	if dbConnection != nil {
		return nil
	}

	config, err := pgxpool.ParseConfig(dbUri)

	if err != nil {
		return err
	}

	if config == nil {
		return errors.New("failed parse config")
	}

	config.MaxConnIdleTime = 5 * time.Minute
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConns = 4

	dbConnection, err = pgxpool.ConnectConfig(context.Background(), config)

	if err != nil {
		return err
	}

	log.Println("Connecting to database...")
	err = dbConnection.Ping(context.Background())

	if err != nil {
		return err
	}
	return nil
}

func QueryRow(query string, args ...interface{}) (pgx.Row, error) {
	if dbConnection == nil {
		return nil, errors.New("database connection is not created")
	}
	return dbConnection.QueryRow(context.Background(), query, args...), nil
}

func QueryRows(query string, args ...interface{}) (pgx.Rows, error) {
	if dbConnection == nil {
		return nil, errors.New("database connection is not created")
	}
	return dbConnection.Query(context.Background(), query, args...)
}

func IncrementExec(query string, args ...interface{}) (int, error) {
	if dbConnection == nil {
		return 0, errors.New("database connection is not created")
	}

	var id int
	row := dbConnection.QueryRow(context.Background(), query, args...)
	err := ScanRow(row, &id)

	return id, err
}

func Exec(query string, args ...interface{}) (int, error) {
	if dbConnection == nil {
		return 0, errors.New("database connection is not created")
	}
	res, err := dbConnection.Exec(context.Background(), query, args...)

	if res == nil {
		return 0, err
	}
	return int(res.RowsAffected()), err
}

func BulkInsert(table string, colums []string, rows [][]interface{}) (int, error) {
	copyCount, err := dbConnection.CopyFrom(
		context.Background(),
		pgx.Identifier{table},
		colums,
		pgx.CopyFromRows(rows),
	)
	return int(copyCount), err
}

func getNullableReplacement(item interface{}) interface{} {
	switch item.(type) {
	case *int:
		return &pgtype.Int4{}
	case *int32:
		return &pgtype.Int4{}
	case *uint32:
		return &pgtype.Int4{}
	case *int64:
		return &pgtype.Int8{}
	case *uint64:
		return &pgtype.Int8{}
	case *bool:
		return &pgtype.Bool{}
	case *float32:
		return &pgtype.Numeric{}
	case *float64:
		return &pgtype.Numeric{}
	case *string:
		return &pgtype.Text{}
	case *time.Time:
		return &pgtype.Timestamp{}
	case *[]string:
		return &pgtype.TextArray{}
	case *[]int:
		return &pgtype.Int4Array{}
	case *[]int32:
		return &pgtype.Int4Array{}
	case *[]int64:
		return &pgtype.Int8Array{}
	}
	return item
}

func getNullableReplacementValue(item interface{}) interface{} {
	switch v := item.(type) {
	case *pgtype.Int4:
		return v.Int
	case *pgtype.Int8:
		return v.Int
	case *pgtype.Bool:
		return v.Bool
	case *pgtype.Float4:
		return v.Float
	case *pgtype.Float8:
		return v.Float
	case *pgtype.Numeric:
		if v.Int != nil {
			return float64(v.Int.Int64()) * math.Pow10(int(v.Exp))
		}
		return float64(0)
	case *pgtype.Text:
		return v.String
	case *pgtype.Timestamp:
		return v.Time

	case *pgtype.Int4Array:
		int4s := make([]int, 0, len(v.Elements))
		for _, element := range v.Elements {
			int4s = append(int4s, int(element.Int))
		}
		return int4s
	case *pgtype.Int8Array:
		int8s := make([]int64, 0, len(v.Elements))
		for _, element := range v.Elements {
			int8s = append(int8s, int64(element.Int))
		}
		return int8s
	case *pgtype.TextArray:
		strs := make([]string, 0, len(v.Elements))
		for _, element := range v.Elements {
			strs = append(strs, element.String)
		}
		return strs
	}
	return nil
}

func updateReplacementValue(item interface{}, replacement interface{}) {
	switch replacement.(type) {
	case *pgtype.Int4:
		switch item.(type) {
		case *int:
			*(item.(*int)) = int(getNullableReplacementValue(replacement).(int32))
			break
		case *int32:
			*(item.(*int32)) = int32(getNullableReplacementValue(replacement).(int32))
			break
		case *uint32:
			*(item.(*uint32)) = uint32(getNullableReplacementValue(replacement).(int32))
			break
		}
	case *pgtype.Int8:
		switch item.(type) {
		case *int64:
			*(item.(*int64)) = int64(getNullableReplacementValue(replacement).(int64))
			break
		case *uint64:
			*(item.(*uint64)) = uint64(getNullableReplacementValue(replacement).(int64))
			break
		}
	case *pgtype.Numeric:
		switch item.(type) {
		case *float32:
			*(item.(*float32)) = float32(getNullableReplacementValue(replacement).(float64))
			break
		case *float64:
			*(item.(*float64)) = float64(getNullableReplacementValue(replacement).(float64))
			break
		}
		break
	case *pgtype.Text:
		switch item.(type) {
		case *string:
			*(item.(*string)) = getNullableReplacementValue(replacement).(string)
			break
		}
		break
	case *pgtype.Timestamp:
		switch item.(type) {
		case *time.Time:
			*(item.(*time.Time)) = getNullableReplacementValue(replacement).(time.Time)
			break
		}
		break
	case *pgtype.Bool:
		switch item.(type) {
		case *bool:
			*(item.(*bool)) = getNullableReplacementValue(replacement).(bool)
			break
		}
		break
	case *pgtype.TextArray:
		switch item.(type) {
		case *[]string:
			*(item.(*[]string)) = getNullableReplacementValue(replacement).([]string)
			break
		}
		break
	case *pgtype.Int4Array:
		switch item.(type) {
		case *[]int:
			*(item.(*[]int)) = getNullableReplacementValue(replacement).([]int)
			break
		}
		break
	case *pgtype.Int8Array:
		switch item.(type) {
		case *[]int64:
			*(item.(*[]int64)) = getNullableReplacementValue(replacement).([]int64)
			break
		}
		break
	}
}

func ScanRow(row pgx.Row, elem ...interface{}) error {

	if len(elem) == 0 {
		return row.Scan(elem...)
	}

	nullables := make([]interface{}, 0, len(elem))

	for i, _ := range elem {
		nullables = append(nullables, getNullableReplacement(elem[i]))
	}

	err := row.Scan(nullables...)

	for i, _ := range elem {
		updateReplacementValue(elem[i], nullables[i])
	}

	return err
}

func ScanRows(rows pgx.Rows, elem ...interface{}) error {

	if len(elem) == 0 {
		return rows.Scan(elem...)
	}

	nullables := make([]interface{}, 0, len(elem))

	for i, _ := range elem {
		nullables = append(nullables, getNullableReplacement(elem[i]))
	}

	err := rows.Scan(nullables...)

	for i, _ := range elem {
		updateReplacementValue(elem[i], nullables[i])
	}

	return err
}

func ToSQLTime(tm time.Time) pgtype.Timestamp {
	ts := pgtype.Timestamp{Time: tm, Status: pgtype.Present}
	if tm.IsZero() {
		ts.Status = pgtype.Null
	}
	return ts
}

func Nullable(item interface{}) interface{} {
	switch item.(type) {
	case int, int32, uint32:
		val := pgtype.Int4{Int: item.(int32), Status: pgtype.Present}
		if item.(int32) == 0 {
			val.Status = pgtype.Null
		}
		return val
	case int64, uint64:
		val := pgtype.Int8{Int: item.(int64), Status: pgtype.Present}
		if item.(int64) == 0 {
			val.Status = pgtype.Null
		}
		return val
	case float32, float64:
		val := pgtype.Numeric{}
		if item.(float64) == 0 {
			val.Status = pgtype.Null
		}
		_ = val.Scan(item)
		return val
	case string:
		val := pgtype.Text{String: item.(string), Status: pgtype.Present}
		if item.(string) == "" {
			val.Status = pgtype.Null
		}
		return val
	case time.Time:
		val := pgtype.Time{Microseconds: item.(time.Time).UnixMicro(), Status: pgtype.Present}
		if item.(time.Time).IsZero() {
			val.Status = pgtype.Null
		}
		return val
	}
	return item
}
