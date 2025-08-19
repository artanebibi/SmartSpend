package database

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"os"
	"path/filepath"
)

func RunMigrations(db *sql.DB) {
	var migrationsPath string

	if os.Getenv("BUILD_ENV") == "docker" {
		migrationsPath = "/app/internal/database/migrations"
	} else {
		var err error = nil
		migrationsPath, err = filepath.Abs("./internal/database/migrations")
		if err != nil {
			log.Fatalf("Failed to get absolute path: %v", err)
		}
	}

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		//pwd, _ := os.Getwd()
		//dir, _ := os.ReadDir(fmt.Sprintf("%s/main", pwd))
		//log.Fatalf("You are here: %s, and the dir\n%s", pwd, dir)

		log.Fatalf("Migrations directory not found at: %s", migrationsPath)
	}

	// Create POSTGRES instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"smartspend",
		driver,
	)
	if err != nil {
		log.Fatalf("Failed to initialize migrations: %v", err)
	}
	if m == nil {
		log.Fatal("Migration instance is nil")
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("\nMigrations completed successfully")
}
