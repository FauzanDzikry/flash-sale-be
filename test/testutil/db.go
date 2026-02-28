package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	pgmod "github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupTestDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	ctx := context.Background()

	postgresContainer, err := pgmod.Run(ctx, "postgres:16-alpine",
		pgmod.WithDatabase("testdb"),
		pgmod.WithUsername("test"),
		pgmod.WithPassword("test"),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get postgres connection string: %v", err)
	}

	var migrateErr error
	for i := 0; i < 10; i++ {
		migrateErr = runMigrations(connStr)
		if migrateErr == nil {
			break
		}
		if i < 9 {
			time.Sleep(500 * time.Millisecond)
		}
	}
	if migrateErr != nil {
		t.Fatalf("failed to run migrations: %v", migrateErr)
	}

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}

	cleanup := func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	}

	return db, cleanup
}

func runMigrations(connStr string) error {
	root := findProjectRoot()
	if root == "" {
		if cwd, err := os.Getwd(); err == nil {
			root = cwd
		} else {
			root = "."
		}
	}
	migrationsPath := filepath.Join(root, "migrations")
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		absPath = migrationsPath
	}
	sourceURL := "file:" + filepath.ToSlash(absPath)

	m, err := migrate.New(sourceURL, connStr)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

func findProjectRoot() string {
	dir, err := filepath.Abs(".")
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "migrations")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
