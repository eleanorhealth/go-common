package infra

import (
	"context"
	"net"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/eleanorhealth/go-common/pkg/env"
	"github.com/eleanorhealth/go-common/pkg/errs"
	"github.com/jackc/pgx/v5"
)

// CloudSQLDialerOption returns a DBOption that wires up the Cloud SQL connector
// dialer when CLOUD_SQL_INSTANCE is set. When the env var is absent the option
// is a no-op, which is the normal path for local dev.
func CloudSQLDialerOption(ctx context.Context) (DBOption, error) {
	cloudSQLInstance := env.Get("CLOUD_SQL_INSTANCE", "")
	if len(cloudSQLInstance) == 0 {
		return func(*pgx.ConnConfig) {}, nil
	}

	d, err := cloudsqlconn.NewDialer(ctx, cloudsqlconn.WithDefaultDialOptions(
		cloudsqlconn.WithPrivateIP(),
	))
	if err != nil {
		return nil, errs.Wrap(err, "initializing Cloud SQL connection dialer")
	}

	return func(config *pgx.ConnConfig) {
		config.Config.DialFunc = func(ctx context.Context, _ string, _ string) (net.Conn, error) {
			return d.Dial(ctx, cloudSQLInstance)
		}
	}, nil
}
