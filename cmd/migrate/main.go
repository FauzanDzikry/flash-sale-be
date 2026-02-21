package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	var (
		up            = flag.Bool("up", false, "run all pending migrations up")
		down          = flag.Bool("down", false, "run migrations down")
		downAll       = flag.Bool("down-all", false, "rollback all migrations (same as -down -steps 0 or max)")
		steps         = flag.Int("steps", 0, "number of migrations to run (with -up or -down). 0 = all")
		force         = flag.Int("force", -1, "force set schema version (use when dirty state). value -1 = disabled")
		dsn           = flag.String("dsn", "", "database URL (overrides DATABASE_URL env)")
		migrationsDir = flag.String("migrations-dir", "", "path to migrations folder (overrides MIGRATIONS_DIR env)")
	)
	flag.Parse()

	dbURL := *dsn
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		fmt.Fprintln(os.Stderr, "error: need DATABASE_URL or -dsn")
		os.Exit(1)
	}

	dir := *migrationsDir
	if dir == "" {
		dir = os.Getenv("MIGRATIONS_DIR")
	}
	if dir == "" {
		dir = "migrations"
	}
	// Di Windows file:///C:/path sering error di driver; pakai path relatif dari project root
	if !filepath.IsAbs(dir) {
		root := findProjectRoot()
		if root != "" {
			_ = os.Chdir(root)
		}
	}
	// Path relatif dari cwd (setelah Chdir ke project root) agar driver file tidak error di Windows
	if dir == "migrations" {
		dir = "./migrations"
	}
	sourceURL := "file://" + filepath.ToSlash(dir)

	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "migrate new:", err)
		os.Exit(1)
	}
	defer m.Close()

	if *force >= 0 {
		if err := m.Force(*force); err != nil {
			fmt.Fprintln(os.Stderr, "force:", err)
			os.Exit(1)
		}
		fmt.Println("forced version to", *force)
		return
	}

	if *up {
		if *steps > 0 {
			err = m.Steps(*steps)
		} else {
			err = m.Up()
		}
		if err != nil && err != migrate.ErrNoChange {
			fmt.Fprintln(os.Stderr, "up:", err)
			os.Exit(1)
		}
		if err == migrate.ErrNoChange {
			fmt.Println("no change (already up to date)")
		} else {
			fmt.Println("migrate up ok")
		}
		return
	}

	if *down || *downAll {
		if *downAll {
			var count int
			for {
				err = m.Steps(-1)
				if err == migrate.ErrNoChange {
					if count == 0 {
						fmt.Println("no change (already at version 0)")
					} else {
						fmt.Printf("migrate down all ok (%d step(s))\n", count)
					}
					break
				}
				if err != nil {
					fmt.Fprintln(os.Stderr, "down:", err)
					os.Exit(1)
				}
				count++
			}
		} else {
			n := *steps
			if n <= 0 {
				n = 1
			}
			err = m.Steps(-n)
			if err != nil && err != migrate.ErrNoChange {
				fmt.Fprintln(os.Stderr, "down:", err)
				os.Exit(1)
			}
			if err == migrate.ErrNoChange {
				fmt.Println("no change (already at version 0)")
			} else {
				fmt.Println("migrate down ok")
			}
		}
		return
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		fmt.Fprintln(os.Stderr, "version:", err)
		os.Exit(1)
	}
	if err == migrate.ErrNilVersion {
		fmt.Println("current version: none")
	} else {
		fmt.Printf("current version: %d (dirty: %v)\n", version, dirty)
	}
}

// findProjectRoot returns directory yang berisi folder "migrations" (naik dari cwd).
func findProjectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := cwd
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
