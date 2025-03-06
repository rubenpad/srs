package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/rubenpad/stock-rating-system/internal/infrastructure/server"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	// Server configuration
	Host            string        `default:"localhost"`
	Port            uint          `default:"8080"`
	ShutdownTimeout time.Duration `default:"10s"`
	// Database configuration
	DatabaseUrl string
}

func Run() error {
	var configuration config
	err := envconfig.Process("SRI", &configuration)
	if err != nil {
		return err
	}

	connectionPoolConfig, err := pgxpool.ParseConfig(configuration.DatabaseUrl + "?sslmode=require&pool_max_conns=40&pool_max_conn_lifetime=300s&pool_max_conn_lifetime_jitter=30s")
	if err != nil {
		return fmt.Errorf("failed to parse connection pool config: %v", err)
	}

	connectionPool, err := pgxpool.NewWithConfig(context.Background(), connectionPoolConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %v", err)
	}

	defer connectionPool.Close()

	ctx, srv := server.New(context.Background(), configuration.Host, configuration.Port, configuration.ShutdownTimeout, connectionPool)

	return srv.Run(ctx)
}
