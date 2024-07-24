package db

import (
	"context"
	"errors"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v4/stdlib"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type DBase struct {
	connection   *pgxpool.Pool
	connMutex    sync.Mutex
	queryTimeout time.Duration
	waitGroup    sync.WaitGroup
}

func (d *DBase) GetConnection() *pgxpool.Pool {
	d.connMutex.Lock()
	defer d.connMutex.Unlock()
	return d.connection
}

func (d *DBase) Close() {
	d.connMutex.Lock()
	defer d.connMutex.Unlock()

	if d.connection == nil {
		return
	}
	d.connection.Close()
	d.connection = nil
}

func (d *DBase) GetStats() *pgxpool.Stat {
	if d.connection == nil {
		return &pgxpool.Stat{}
	}
	return d.connection.Stat()
}

func Create(dbUri string, queryTimeout time.Duration) (*DBase, error) {

	if dbUri == "" {
		return nil, errors.New("empty database URI")
	}

	var err error
	db := DBase{
		connection:   nil,
		queryTimeout: queryTimeout,
	}

	config, err := pgxpool.ParseConfig(dbUri)

	if err != nil {
		return nil, err
	}

	if config == nil {
		return nil, errors.New("failed parse config")
	}

	config.MaxConnIdleTime = 5 * time.Minute
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConns = 4

	db.connection, err = pgxpool.ConnectConfig(context.Background(), config)

	if err != nil {
		return nil, err
	}

	log.Debugln("Connecting to DB...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = db.connection.Ping(ctx)

	if err != nil {
		return nil, err
	}
	log.Println("DB connected")
	return &db, nil
}

func (d *DBase) QueryRow(query string, args ...interface{}) (pgx.Row, error) {
	if d.connection == nil {
		return nil, errors.New("database connection is not created")
	}
	ctx, cancel := context.WithTimeout(context.Background(), d.queryTimeout)
	defer cancel()

	d.waitGroup.Add(1)
	defer d.waitGroup.Done()

	return d.connection.QueryRow(ctx, query, args...), nil
}

func (d *DBase) QueryRows(query string, args ...interface{}) (pgx.Rows, error) {
	if d.connection == nil {
		return nil, errors.New("database connection is not created")
	}
	ctx, cancel := context.WithTimeout(context.Background(), d.queryTimeout)
	defer cancel()

	d.waitGroup.Add(1)
	defer d.waitGroup.Done()

	return d.connection.Query(ctx, query, args...)
}

func (d *DBase) IncrementExec(query string, args ...interface{}) (int, error) {
	if d.connection == nil {
		return 0, errors.New("database connection is not created")
	}

	var id int

	ctx, cancel := context.WithTimeout(context.Background(), d.queryTimeout)
	defer cancel()

	d.waitGroup.Add(1)
	defer d.waitGroup.Done()

	row := d.connection.QueryRow(ctx, query, args...)
	err := ScanRow(row, &id)

	return id, err
}

func (d *DBase) Exec(query string, args ...interface{}) (int, error) {
	if d.connection == nil {
		return 0, errors.New("database connection is not created")
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.queryTimeout)
	defer cancel()

	d.waitGroup.Add(1)
	defer d.waitGroup.Done()

	res, err := d.connection.Exec(ctx, query, args...)

	if res == nil {
		return 0, err
	}
	return int(res.RowsAffected()), err
}

func (d *DBase) BulkInsert(table string, columns []string, rows [][]interface{}) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.queryTimeout)
	defer cancel()

	copyCount, err := d.connection.CopyFrom(
		ctx,
		pgx.Identifier{table},
		columns,
		pgx.CopyFromRows(rows),
	)
	return int(copyCount), err
}

func (d *DBase) WaitAllComplete() {
	d.waitGroup.Wait()
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
	case *[]uint32:
		return &pgtype.Int4Array{}
	case *[]int64:
		return &pgtype.Int8Array{}
	case *[]uint64:
		return &pgtype.Int8Array{}
	}
	return item
}

func getNullableReplacementValue(item interface{}, variant interface{}) interface{} {
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
			return NumericGetFloat64(v)
		}
		return float64(0)
	case *pgtype.Text:
		return v.String
	case *pgtype.Timestamp:
		return v.Time
		/* Skip by auto replacement, we use only pgtype.Timestamp<->pgtype.Time
		case *pgtype.Timestamptz:
			return v.Time
		case *pgtype.Time:
			return time.UnixMicro(v.Microseconds)
		*/

	case *pgtype.Int4Array:
		switch variant.(type) {
		case *[]int:
			int4s := make([]int, 0, len(v.Elements))
			for _, element := range v.Elements {
				int4s = append(int4s, int(element.Int))
			}
			return int4s
		case *[]int32:
			int4s := make([]int32, 0, len(v.Elements))
			for _, element := range v.Elements {
				int4s = append(int4s, int32(element.Int))
			}
			return int4s
		case *[]uint32:
			int4s := make([]uint32, 0, len(v.Elements))
			for _, element := range v.Elements {
				int4s = append(int4s, uint32(element.Int))
			}
			return int4s
		}
		return nil
	case *pgtype.Int8Array:
		switch variant.(type) {
		case *[]int64:
			int8s := make([]int64, 0, len(v.Elements))
			for _, element := range v.Elements {
				int8s = append(int8s, int64(element.Int))
			}
			return int8s
		case *[]uint64:
			int8s := make([]uint64, 0, len(v.Elements))
			for _, element := range v.Elements {
				int8s = append(int8s, uint64(element.Int))
			}
			return int8s
		}
		return nil
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
			*(item.(*int)) = int(getNullableReplacementValue(replacement, item).(int32))
			break
		case *int32:
			*(item.(*int32)) = int32(getNullableReplacementValue(replacement, item).(int32))
			break
		case *uint32:
			*(item.(*uint32)) = uint32(getNullableReplacementValue(replacement, item).(int32))
			break
		}
	case *pgtype.Int8:
		switch item.(type) {
		case *int64:
			*(item.(*int64)) = int64(getNullableReplacementValue(replacement, item).(int64))
			break
		case *uint64:
			*(item.(*uint64)) = uint64(getNullableReplacementValue(replacement, item).(int64))
			break
		}
	case *pgtype.Numeric:
		switch item.(type) {
		case *float32:
			*(item.(*float32)) = float32(getNullableReplacementValue(replacement, item).(float64))
			break
		case *float64:
			*(item.(*float64)) = float64(getNullableReplacementValue(replacement, item).(float64))
			break
		}
		break
	case *pgtype.Text:
		switch item.(type) {
		case *string:
			*(item.(*string)) = getNullableReplacementValue(replacement, item).(string)
			break
		}
		break
	case *pgtype.Timestamp:
		switch item.(type) {
		case *time.Time:
			*(item.(*time.Time)) = getNullableReplacementValue(replacement, item).(time.Time)
			break
		}
		break
	case *pgtype.Timestamptz:
		switch item.(type) {
		case *time.Time:
			*(item.(*time.Time)) = getNullableReplacementValue(replacement, item).(time.Time)
			break
		}
		break
	case *pgtype.Time:
		switch item.(type) {
		case *time.Time:
			*(item.(*time.Time)) = getNullableReplacementValue(replacement, item).(time.Time)
			break
		}
		break
	case *pgtype.Bool:
		switch item.(type) {
		case *bool:
			*(item.(*bool)) = getNullableReplacementValue(replacement, item).(bool)
			break
		}
		break
	case *pgtype.TextArray:
		switch item.(type) {
		case *[]string:
			*(item.(*[]string)) = getNullableReplacementValue(replacement, item).([]string)
			break
		}
		break
	case *pgtype.Int4Array:
		switch item.(type) {
		case *[]int:
			*(item.(*[]int)) = getNullableReplacementValue(replacement, item).([]int)
			break
		case *[]int32:
			*(item.(*[]int32)) = getNullableReplacementValue(replacement, item).([]int32)
			break
		case *[]uint32:
			*(item.(*[]uint32)) = getNullableReplacementValue(replacement, item).([]uint32)
			break
		}
		break
	case *pgtype.Int8Array:
		switch item.(type) {
		case *[]int64:
			*(item.(*[]int64)) = getNullableReplacementValue(replacement, item).([]int64)
			break
		case *[]uint64:
			*(item.(*[]uint64)) = getNullableReplacementValue(replacement, item).([]uint64)
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

	for i := range elem {
		nullables = append(nullables, getNullableReplacement(elem[i]))
	}

	err := row.Scan(nullables...)

	for i := range elem {
		updateReplacementValue(elem[i], nullables[i])
	}

	return err
}

func ScanRows(rows pgx.Rows, elem ...interface{}) error {

	if len(elem) == 0 {
		return rows.Scan(elem...)
	}

	nullables := make([]interface{}, 0, len(elem))

	for i := range elem {
		nullables = append(nullables, getNullableReplacement(elem[i]))
	}

	err := rows.Scan(nullables...)

	for i := range elem {
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
	case int64:
		val := pgtype.Int8{Int: item.(int64), Status: pgtype.Present}
		_ = val.Set(item)
		if item.(int64) == 0 {
			val.Status = pgtype.Null
		}
		return val
	case uint64:
		val := pgtype.Int8{Int: int64(item.(uint64)), Status: pgtype.Present}
		_ = val.Set(item)
		if item.(uint64) == 0 {
			val.Status = pgtype.Null
		}
		return val
	case int32:
		val := pgtype.Int4{Int: item.(int32), Status: pgtype.Present}
		if item.(int32) == 0 {
			val.Status = pgtype.Null
		}
		return val
	case uint32:
		val := pgtype.Int4{Int: int32(item.(uint32)), Status: pgtype.Present}
		if item.(uint32) == 0 {
			val.Status = pgtype.Null
		}
		return val
	case int:
		val := pgtype.Int4{Int: int32(item.(int)), Status: pgtype.Present}
		if item.(int) == 0 {
			val.Status = pgtype.Null
		}
		return val
	case []int32, []uint32, []int:
		val := pgtype.Int4Array{}
		_ = val.Set(item)
		if item == nil {
			val.Status = pgtype.Null
		}
		return val
	case []int64, []uint64:
		val := pgtype.Int8Array{}
		_ = val.Set(item)
		if item == nil {
			val.Status = pgtype.Null
		}
		return val
	case float32:
		val := pgtype.Numeric{}
		_ = val.Set(item)
		if item.(float32) == 0 {
			val.Status = pgtype.Null
		}
		return val
	case float64:
		val := pgtype.Numeric{}
		_ = val.Set(item)
		if item.(float64) == 0 {
			val.Status = pgtype.Null
		}
		return val
	case string:
		val := pgtype.Text{}
		_ = val.Set(item)
		if item.(string) == "" {
			val.Status = pgtype.Null
		}
		return val
	case []string:
		val := pgtype.TextArray{}
		_ = val.Set(item)
		if len(item.([]string)) == 0 {
			val.Status = pgtype.Null
		}
		return val
	case time.Time:
		val := pgtype.Timestamptz{Time: item.(time.Time), Status: pgtype.Present}
		if item.(time.Time).IsZero() {
			val.Status = pgtype.Null
		}
		return val
	}
	return item
}

func NumericGetInt64(numeric *pgtype.Numeric) int64 {
	var val int64
	_ = numeric.AssignTo(&val)
	return val
}

func NumericGetFloat64(numeric *pgtype.Numeric) float64 {
	var val float64
	_ = numeric.AssignTo(&val)
	return val
}
