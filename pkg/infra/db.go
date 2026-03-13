package infra

import (
	"context"
	"database/sql"
	"time"

	sqltrace "github.com/DataDog/dd-trace-go/contrib/database/sql/v2"
	"github.com/avast/retry-go"
	"github.com/eleanorhealth/go-common/pkg/env"
	"github.com/eleanorhealth/go-common/pkg/errs"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

const (
	// Default number of max database connections.
	dbMaxOpenConns = 5
)

// DBOption is a functional option applied to the pgx.ConnConfig before the
// sql.DB is opened. Use it to inject custom dial functions, TLS settings, etc.
type DBOption func(*pgx.ConnConfig)

type DBExecutor interface {
	Execute(ctx context.Context, query string, args ...any) (int64, error)
}

type DBQuerier interface {
	Query(ctx context.Context, dst any, query string, args ...any) error
	QueryRow(ctx context.Context, dst any, query string, args ...any) error
}

type DBExecutorQuerier interface {
	DBExecutor
	DBQuerier
}

func DB(ctx context.Context, connString string, traceServiceName string, opts ...DBOption) (*sql.DB, error) {
	config, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, errs.Wrap(err, "parsing connection string")
	}

	for _, opt := range opts {
		opt(config)
	}

	connector := stdlib.GetConnector(*config)
	sqltrace.Register("pgx", stdlib.GetDefaultDriver())
	db := sqltrace.OpenDB(connector, sqltrace.WithService(traceServiceName))

	if v, exists := env.GetExists[int]("DB_CONN_MAX_IDLE_TIME"); exists {
		db.SetConnMaxIdleTime(time.Duration(v) * time.Minute)
	}

	if v, exists := env.GetExists[int]("DB_CONN_MAX_LIFETIME"); exists {
		db.SetConnMaxLifetime(time.Duration(v) * time.Minute)
	}

	if v, exists := env.GetExists[int]("DB_MAX_IDLE_CONNS"); exists {
		db.SetMaxIdleConns(v)
	}

	if v, exists := env.GetExists[int]("DB_MAX_OPEN_CONNS"); exists {
		db.SetMaxOpenConns(v)
	} else {
		db.SetMaxOpenConns(dbMaxOpenConns)
	}

	err = retry.Do(func() error {
		return db.PingContext(ctx)
	})
	if err != nil {
		return nil, errs.Wrap(err, "pinging database")
	}

	return db, nil
}

type SQLExecutorQuerier struct {
	db *sql.DB
}

var (
	_ DBExecutor        = (*SQLExecutorQuerier)(nil)
	_ DBQuerier         = (*SQLExecutorQuerier)(nil)
	_ DBExecutorQuerier = (*SQLExecutorQuerier)(nil)
)

func NewSQLExecutorQuerier(db *sql.DB) *SQLExecutorQuerier {
	return &SQLExecutorQuerier{
		db: db,
	}
}

func (s *SQLExecutorQuerier) Execute(ctx context.Context, query string, args ...any) (int64, error) {
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, errs.Wrap(err, "executing query")
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, errs.Wrap(err, "getting rows affected")
	}

	return affected, nil
}

func (s *SQLExecutorQuerier) Query(ctx context.Context, dst any, query string, args ...any) error {
	err := sqlscan.Select(ctx, s.db, dst, query, args...)
	if err != nil {
		return errs.Wrap(err, "querying and scanning rows")
	}

	return nil
}

func (s *SQLExecutorQuerier) QueryRow(ctx context.Context, dst any, query string, args ...any) error {
	err := sqlscan.Get(ctx, s.db, dst, query, args...)
	if err != nil {
		return errs.Wrap(err, "querying and scanning row")
	}

	return nil
}
