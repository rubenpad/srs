package main

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Database         string `required:"true"`
	DatabaseHost     string `required:"true" split_words:"true"`
	DatabaseUser     string `required:"true" split_words:"true"`
	DatabasePort     uint   `required:"true" split_words:"true"`
	DatabasePassword string `required:"true" split_words:"true"`
}

func main() {
	log.Println("migrations process started")
	start := time.Now()

	var configuration config
	err := envconfig.Process("SRS", &configuration)
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

		if downErr := migration.Down(); downErr != nil {
			log.Fatal("error executing up & down migrations")
		}

		log.Fatal("error executing up migrations", err)
	}

	elapsed := time.Since(start)
	minutes := int(elapsed.Minutes())
	seconds := int(elapsed.Seconds()) % 60
	milliseconds := int(elapsed.Milliseconds()) % 1000
	log.Printf("migrations executed successfully: %dm %ds %dms", minutes, seconds, milliseconds)
}
