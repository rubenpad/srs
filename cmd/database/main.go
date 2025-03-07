package main

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Database         string
	DatabaseHost     string `required:"true" split_words:"true"`
	DatabaseUser     string `required:"true" split_words:"true"`
	DatabasePort     uint   `required:"true" split_words:"true"`
	DatabasePassword string `required:"true" split_words:"true"`
}

func main() {
	var configuration config
	err := envconfig.Process("SRI", &configuration)
	if err != nil {
		log.Fatal("error getting database configuration values")
	}

	connectionParams := "?sslmode=require&pool_max_conns=40&pool_max_conn_lifetime=300s&pool_max_conn_lifetime_jitter=30s"
	connectionString := fmt.Sprintf("cockroachdb://%s:%s@%s:%d/%s", configuration.DatabaseUser, configuration.DatabasePassword, configuration.DatabaseHost, configuration.DatabasePort, configuration.Database) + connectionParams

	migration, err := migrate.New("file://database/migrations", connectionString)

	if err != nil {
		log.Fatal("error configuring migrations", err)
	}

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("error executing migrations", err)
	}

	log.Println("migrations executed successfully")
}
