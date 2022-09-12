package testhelper

import (
	"database/sql"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func NewPgMigrator(db *sql.DB) (*migrate.Migrate, error) {

	// sourceUrl := "file://C:/Users/Asus/Documents/prog/away_backend/testhelper/migrations"
	sourceUrl := "file:/Users/thomas/Documents/projects/away_backend/db/migration"
	if f := os.Getenv("MIGRATIONS_FOLDER"); f != "" {
		sourceUrl = f
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})

	if err != nil {
		log.Fatalf("failed to create migrator driver: %s", err)
	}

	m, err := migrate.NewWithDatabaseInstance(sourceUrl, "postgres", driver)

	return m, err
}
