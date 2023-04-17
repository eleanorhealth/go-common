package infra

import (
	"context"

	"github.com/avast/retry-go"
	"github.com/eleanorhealth/go-common/pkg/errs"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
)

func PgxPool(ctx context.Context, connString string, traceServiceName string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, errs.Wrap(err, "parsing connection string")
	}

	err = setCloudSQLInstanceDialFunc(ctx, config.ConnConfig)
	if err != nil {
		return nil, errs.Wrap(err, "setting Cloud SQL instance dial func")
	}

	// We don't need to worry about setting a default max number of database
	// connections here because pgx defaults to the greater of 4 or runtime.NumCPU().

	conn, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, errs.Wrap(err, "creating pool")
	}

	// Waiting on Go 1.20 for tracing support in pgx: https://github.com/DataDog/dd-trace-go/pull/1537

	err = retry.Do(func() error {
		return conn.Ping(ctx)
	})
	if err != nil {
		return nil, errs.Wrap(err, "pinging database")
	}

	return conn, nil
}

type PgxPoolExecutorQuerier struct {
	pool *pgxpool.Pool
}

var _ DBExecutor = (*PgxPoolExecutorQuerier)(nil)
var _ DBQuerier = (*PgxPoolExecutorQuerier)(nil)
var _ DBExecutorQuerier = (*PgxPoolExecutorQuerier)(nil)

func NewPgxExecutorQuerier(pool *pgxpool.Pool) *PgxPoolExecutorQuerier {
	return &PgxPoolExecutorQuerier{
		pool: pool,
	}
}

func (s *PgxPoolExecutorQuerier) Execute(ctx context.Context, query string, args ...any) (int64, error) {
	cmd, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, errs.Wrap(err, "executing query")
	}

	return cmd.RowsAffected(), nil
}

func (s *PgxPoolExecutorQuerier) Query(ctx context.Context, dst any, query string, args ...any) error {
	err := pgxscan.Select(ctx, s.pool, dst, query, args...)
	if err != nil {
		return errs.Wrap(err, "querying and scanning rows")
	}

	return nil
}

func (s *PgxPoolExecutorQuerier) QueryRow(ctx context.Context, dst any, query string, args ...any) error {
	err := pgxscan.Get(ctx, s.pool, dst, query, args...)
	if err != nil {
		return errs.Wrap(err, "querying and scanning row")
	}

	return nil
}
