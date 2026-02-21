package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
		create        = flag.String("create", "", "create new migration files (e.g. -create add_products_table). no DB needed")
		dsn           = flag.String("dsn", "", "database URL (overrides DATABASE_URL env)")
		migrationsDir = flag.String("migrations-dir", "", "path to migrations folder (overrides MIGRATIONS_DIR env)")
	)
	flag.Parse()

	if *create != "" {
		if err := runCreate(*create, *migrationsDir); err != nil {
			fmt.Fprintln(os.Stderr, "create:", err)
			os.Exit(1)
		}
		fmt.Println("migration files created")
		return
	}

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

// runCreate membuat pasangan file .up.sql dan .down.sql di folder migrations.
// Nama migration disanitasi (spasi -> underscore, lowercase).
func runCreate(name, migrationsDir string) error {
	dir := migrationsDir
	if dir == "" {
		dir = os.Getenv("MIGRATIONS_DIR")
	}
	if dir == "" {
		dir = "migrations"
	}
	root := findProjectRoot()
	if root != "" {
		_ = os.Chdir(root)
	}
	dirPath := dir
	if !filepath.IsAbs(dirPath) {
		if root != "" {
			dirPath = filepath.Join(root, dir)
		} else {
			var err error
			dirPath, err = filepath.Abs(dir)
			if err != nil {
				dirPath = dir
			}
		}
	}
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}
	nextVer, err := nextMigrationVersion(dirPath)
	if err != nil {
		return err
	}
	safeName := sanitizeMigrationName(name)
	baseName := fmt.Sprintf("%06d_%s", nextVer, safeName)
	upPath := filepath.Join(dirPath, baseName+".up.sql")
	downPath := filepath.Join(dirPath, baseName+".down.sql")
	upContent := "-- migration up: " + name + "\n"
	downContent := "-- migration down: " + name + "\n"
	if err := os.WriteFile(upPath, []byte(upContent), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(downPath, []byte(downContent), 0644); err != nil {
		_ = os.Remove(upPath)
		return err
	}
	fmt.Printf("created %s\n", filepath.Join(dir, baseName+".up.sql"))
	fmt.Printf("created %s\n", filepath.Join(dir, baseName+".down.sql"))
	return nil
}

var migrationVersionRe = regexp.MustCompile(`^(\d{6})_.*\.(up|down)\.sql$`)

func nextMigrationVersion(dirPath string) (int, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return 1, nil
	}
	maxVer := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		matches := migrationVersionRe.FindStringSubmatch(e.Name())
		if len(matches) >= 2 {
			var v int
			if _, err := fmt.Sscanf(matches[1], "%d", &v); err == nil && v > maxVer {
				maxVer = v
			}
		}
	}
	return maxVer + 1, nil
}

func sanitizeMigrationName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	name = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")
	if name == "" {
		name = "migration"
	}
	return name
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
