package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func FindMigrationsDir(levels int) string {
	const migrationsDirName = "database/migrations"

	if infoPath, err := os.Stat(migrationsDirName); err == nil && infoPath.IsDir() {
		return migrationsDirName
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for range levels {
		currentDir = filepath.Dir(currentDir)
		migrationsPath := filepath.Join(currentDir, migrationsDirName)
		if infoPath, err := os.Stat(migrationsPath); err == nil && infoPath.IsDir() {
			return migrationsPath
		}
	}

	currentDir, err = os.Getwd()
	if err != nil {
		return ""
	}

	for range levels {
		entries, err := os.ReadDir(currentDir)
		if err != nil {
			return ""
		}

		found := false
		for _, entry := range entries {
			if entry.IsDir() {
				migrationsPath := filepath.Join(currentDir, entry.Name(), migrationsDirName)
				if infoPath, err := os.Stat(migrationsPath); err == nil && infoPath.IsDir() {
					return migrationsPath
				}
			}
		}

		if !found {
			break
		}
		currentDir = filepath.Dir(currentDir)
	}

	return ""
}

func Migrations() {
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s?multiStatements=true",
		user, pass, host, port, name)

	migrationsDir := FindMigrationsDir(3)
	if migrationsDir == "" {
		log.Fatal("Could not find migrations directory")
	}

	m, err := migrate.New(
		"file://"+migrationsDir,
		dsn,
	)
	if err != nil {
		log.Fatal("Migration setup failed:", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Migration failed:", err)
	}

	log.Println("Database Migrated")
}

func MigrateDown() error {
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s?multiStatements=true",
		user, pass, host, port, name)

	migrationsDir := FindMigrationsDir(3)
	if migrationsDir == "" {
		return fmt.Errorf("could not find migrations directory")
	}

	m, err := migrate.New(
		"file://"+migrationsDir,
		dsn,
	)
	if err != nil {
		return fmt.Errorf("migration setup failed: %v", err)
	}

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration down failed: %v", err)
	}

	log.Println("Database Migration Down completed")
	return nil
}
