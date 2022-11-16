package infra

import (
	"context"
	"database/sql"
	"net"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/avast/retry-go"
	"github.com/eleanorhealth/go-common/pkg/env"
	"github.com/eleanorhealth/go-common/pkg/errs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
)

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
	sqltrace.Register("db", stdlib.GetDefaultDriver())
	db := sqltrace.OpenDB(connector, sqltrace.WithServiceName(traceServiceName))

	connMaxIdleTime := env.Get("DB_CONN_MAX_IDLE_TIME", 0)
	connMaxLifetime := env.Get("DB_CONN_MAX_LIFETIME", 0)
	maxIdleConns := env.Get("DB_MAX_IDLE_CONNS", 0)
	maxOpenConns := env.Get("DB_MAX_OPEN_CONNS", 0)

	if connMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(time.Duration(connMaxIdleTime) * time.Minute)
	}

	if connMaxLifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Minute)
	}

	if maxIdleConns > 0 {
		db.SetMaxIdleConns(maxIdleConns)
	}

	if maxOpenConns > 0 {
		db.SetMaxOpenConns(maxOpenConns)
	}

	err = retry.Do(func() error {
		return db.PingContext(ctx)
	})
	if err != nil {
		return nil, errs.Wrap(err, "pinging database")
	}

	return db, nil
}

func PgxPool(ctx context.Context, connString string, traceServiceName string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, errs.Wrap(err, "parsing connection string")
	}

	err = setCloudSQLInstanceDialFunc(ctx, config.ConnConfig)
	if err != nil {
		return nil, errs.Wrap(err, "setting Cloud SQL instance dial func")
	}

	conn, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, errs.Wrap(err, "creating pool")
	}

	// Waiting on tracing support for pgx: https://github.com/DataDog/dd-trace-go/pull/1537

	err = retry.Do(func() error {
		return conn.Ping(ctx)
	})
	if err != nil {
		return nil, errs.Wrap(err, "pinging database")
	}

	return conn, nil
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
