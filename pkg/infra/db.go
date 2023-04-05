package infra

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/avast/retry-go"
	"github.com/eleanorhealth/go-common/pkg/env"
	"github.com/eleanorhealth/go-common/pkg/errs"
	"github.com/georgysavva/scany/sqlscan"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
)

var (
	ErrNoTransaction = errors.New("no transaction")
)

const (
	// Default number of max database connections.
	dbMaxOpenConns = 5
)

type DBExecutor interface {
	Commit() error
	Execute(ctx context.Context, query string, args ...any) (int64, error)
	Rollback() error
	Transaction(ctx context.Context) (DBExecutor, error)
}

type DBQuerier interface {
	Query(ctx context.Context, dst any, query string, args ...any) error
	QueryRow(ctx context.Context, dst any, query string, args ...any) error
}

type DBExecutorQuerier interface {
	DBExecutor
	DBQuerier
}

func DB(ctx context.Context, connString string, traceServiceName string) (*sql.DB, error) {
	config, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, errs.Wrap(err, "parsing connection string")
	}

	err = setCloudSQLInstanceDialFunc(ctx, config)
	if err != nil {
		return nil, errs.Wrap(err, "setting Cloud SQL instance dial func")
	}

	connector := stdlib.GetConnector(*config)
	sqltrace.Register("pgx", stdlib.GetDefaultDriver())
	db := sqltrace.OpenDB(connector, sqltrace.WithServiceName(traceServiceName))

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

func setCloudSQLInstanceDialFunc(ctx context.Context, config *pgx.ConnConfig) error {
	cloudSQLInstance := env.Get("CLOUD_SQL_INSTANCE", "")
	if len(cloudSQLInstance) > 0 {
		d, err := cloudsqlconn.NewDialer(ctx, cloudsqlconn.WithDefaultDialOptions(
			cloudsqlconn.WithPrivateIP(),
		))
		if err != nil {
			return errs.Wrap(err, "initializing connection dialer")
		}

		config.DialFunc = func(ctx context.Context, _ string, instance string) (net.Conn, error) {
			return d.Dial(ctx, cloudSQLInstance)
		}
	}

	return nil
}

type SQLExecutorQuerier struct {
	db *sql.DB
	tx *sql.Tx
}

var _ DBExecutor = (*SQLExecutorQuerier)(nil)
var _ DBQuerier = (*SQLExecutorQuerier)(nil)
var _ DBExecutorQuerier = (*SQLExecutorQuerier)(nil)

func NewSQLExecutorQuerier(db *sql.DB, tx *sql.Tx) *SQLExecutorQuerier {
	return &SQLExecutorQuerier{
		db: db,
		tx: tx,
	}
}

func (s *SQLExecutorQuerier) Execute(ctx context.Context, query string, args ...any) (int64, error) {
	var res sql.Result
	var err error

	if s.tx != nil {
		res, err = s.tx.ExecContext(ctx, query, args...)
		if err != nil {
			return 0, errs.Wrap(err, "executing tx query")
		}
	} else {
		res, err = s.db.ExecContext(ctx, query, args...)
		if err != nil {
			return 0, errs.Wrap(err, "executing query")
		}
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, errs.Wrap(err, "getting rows affected")
	}

	return affected, nil
}

func (s *SQLExecutorQuerier) Transaction(ctx context.Context) (DBExecutor, error) {
	if s.tx != nil {
		return s, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errs.Wrap(err, "beginning tx")
	}

	s.tx = tx

	return s, nil
}

func (s *SQLExecutorQuerier) Commit() error {
	if s.tx == nil {
		return errs.Wrap(ErrNoTransaction, "commit")
	}

	err := s.tx.Commit()

	s.tx = nil

	return err
}

func (s *SQLExecutorQuerier) Rollback() error {
	if s.tx == nil {
		return errors.New("no tx in progress to rollback")
	}

	err := s.tx.Rollback()

	s.tx = nil

	return err
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
