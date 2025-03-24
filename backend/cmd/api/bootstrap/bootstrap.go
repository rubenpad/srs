package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
	"github.com/rubenpad/srs/internal/infrastructure/logging"
	"github.com/rubenpad/srs/internal/infrastructure/otel"
	"github.com/rubenpad/srs/internal/infrastructure/server"
)

type config struct {
	// Server configuration
	ShutdownTimeout time.Duration `default:"10s"`
	// Database configuration
	Database         string
	DatabaseHost     string `required:"true" split_words:"true"`
	DatabaseUser     string `required:"true" split_words:"true"`
	DatabasePort     uint   `required:"true" split_words:"true"`
	DatabasePassword string `required:"true" split_words:"true"`
}

func Run() error {
	logging.Set()

	var configuration config
	err := envconfig.Process("SRS", &configuration)
	if err != nil {
		return err
	}

	tp, tracerError := otel.InitTracer()
	if tracerError != nil {
		return err
	}

	defer func() {
		if tpError := tp.Shutdown(context.Background()); tpError != nil {
			slog.Error("error shutting down tracer", "error", tpError.Error())
		}
	}()

	connectionParams := "?sslmode=require&pool_max_conns=40&pool_max_conn_lifetime=300s&pool_max_conn_lifetime_jitter=30s"
	connectionString := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", configuration.DatabaseUser, configuration.DatabasePassword, configuration.DatabaseHost, configuration.DatabasePort, configuration.Database) + connectionParams

	connectionPoolContext := context.Background()
	connectionPool, configError := pgxpool.New(connectionPoolContext, connectionString)

	if err := connectionPool.Ping(connectionPoolContext); err != nil || configError != nil {
		return fmt.Errorf("failed to create connection pool: %v", err)
	}

	defer connectionPool.Close()

	ctx, srv := server.New(context.Background(), "0.0.0.0", 8080, configuration.ShutdownTimeout, connectionPool)

	return srv.Run(ctx)
}
