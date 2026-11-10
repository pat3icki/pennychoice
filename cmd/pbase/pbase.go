package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Define command-line flags
	testFlag := flag.Bool("test", false, "Use test environment with db.env file")
	helpFlag := flag.Bool("help", false, "Show help message")

	flag.Parse()

	if *helpFlag {
		printHelp()
		return
	}

	// Get the command (up, down, etc.)
	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("Error: command required (e.g., 'up', 'down', 'version')")
	}

	command := args[0]

	// Determine environment
	var dbURL string
	var err error

	if *testFlag {
		dbURL, err = getTestDBURL()
		if err != nil {
			log.Fatalf("Failed to get test database URL: %v", err)
		}
		log.Println("Running migrations in TEST environment")
	} else {
		// Default: use environment variable or hardcoded URL
		dbURL = os.Getenv("DATABASE_URL")
		if dbURL == "" {
			log.Fatal("DATABASE_URL environment variable not set. Use -test flag for test environment.")
		}
		log.Println("Running migrations in production environment")
	}

	// Run the migration command
	if err := runMigration(dbURL, command, args[1:]); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}

// printHelp displays usage information
func printHelp() {
	fmt.Println("Usage: pbase migrate [options] <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up               Apply all pending migrations")
	fmt.Println("  up N             Apply N pending migrations")
	fmt.Println("  down             Rollback the last migration")
	fmt.Println("  down N           Rollback N migrations")
	fmt.Println("  down -all        Rollback all migrations")
	fmt.Println("  version          Show current migration version")
	fmt.Println("  force VERSION    Force set version (use with caution)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -test            Use test environment with tests/db.env file")
	fmt.Println("  -help            Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  pbase migrate up")
	fmt.Println("  pbase migrate -test up")
	fmt.Println("  pbase migrate up 2")
	fmt.Println("  pbase migrate down")
	fmt.Println("  pbase migrate version")
}

// getTestDBURL reads the db.env file and constructs the database URL
func getTestDBURL() (string, error) {
	root := findProjectRoot()
	envPath := filepath.Join(root, "tests", "db.env")

	file, err := os.Open(envPath)
	if err != nil {
		return "", fmt.Errorf("failed to open env file %s: %w", envPath, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read env file: %w", err)
	}

	reader := bytes.NewReader(data)
	envMap, err := godotenv.Parse(reader)
	if err != nil {
		return "", fmt.Errorf("failed to parse env file: %w", err)
	}

	// Build the database URL from env variables
	dbURL := buildDBURL(envMap)
	if dbURL == "" {
		return "", fmt.Errorf("missing required database environment variables")
	}

	return dbURL, nil
}

// buildDBURL constructs the database connection string from env map
func buildDBURL(envMap map[string]string) string {
	// Check if DATABASE_URL is directly provided
	if url, ok := envMap["DATABASE_URL"]; ok && url != "" {
		return url
	}

	// Build from individual components
	host := getEnv(envMap, "DB_HOST", "localhost")
	port := getEnv(envMap, "DB_PORT", "5432")
	user := getEnv(envMap, "DB_USER", "")
	password := getEnv(envMap, "DB_PASSWORD", "")
	dbname := getEnv(envMap, "DB_NAME", "")
	sslmode := getEnv(envMap, "DB_SSLMODE", "require")

	if user == "" || dbname == "" {
		return ""
	}

	// Build the connection string with SSL
	if password != "" {
		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			user, password, host, port, dbname, sslmode)
	}
	return fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s",
		user, host, port, dbname, sslmode)
}

// getEnv gets environment variable with fallback
func getEnv(envMap map[string]string, key, fallback string) string {
	if value, ok := envMap[key]; ok && value != "" {
		return value
	}
	return fallback
}

// findProjectRoot finds the project root by looking for go.mod
func findProjectRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("cannot get caller info")
	}
	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			log.Fatal("go.mod not found; are you inside the module?")
		}
		dir = parent
	}
}

// runMigration executes the migration command
func runMigration(dbURL, command string, args []string) error {
	// Determine migration path
	root := findProjectRoot()
	migrationPath := filepath.Join(root, "database", "migration")

	// Check if migration directory exists
	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		return fmt.Errorf("migration directory not found at %s", migrationPath)
	}

	// Create migration instance
	m, err := createMigrationInstance(dbURL, migrationPath)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	// Execute the command
	switch command {
	case "up":
		return runUp(m, args)
	case "down":
		return runDown(m, args)
	case "version":
		return runVersion(m)
	case "force":
		return runForce(m, args)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// createMigrationInstance creates a new migrate instance
func createMigrationInstance(dbURL, migrationPath string) (*migrate.Migrate, error) {
	// Open database connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not ping database: %w", err)
	}

	// Create postgres driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not create database driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationPath,
		"postgres",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create migrate instance: %w", err)
	}

	return m, nil
}

// runUp executes the up command
func runUp(m *migrate.Migrate, args []string) error {
	if len(args) == 0 {
		log.Println("Applying all pending migrations...")
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("up migration failed: %w", err)
		}
		log.Println("All migrations applied successfully")
		return nil
	}

	// Try to parse N as integer
	var n int
	if _, err := fmt.Sscanf(args[0], "%d", &n); err != nil {
		return fmt.Errorf("invalid number of migrations: %s", args[0])
	}

	log.Printf("Applying %d pending migrations...\n", n)
	if err := m.Steps(n); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("up migration failed: %w", err)
	}
	log.Printf("Successfully applied %d migrations", n)
	return nil
}

// runDown executes the down command
func runDown(m *migrate.Migrate, args []string) error {
	if len(args) == 0 {
		log.Println("Rolling back last migration...")
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("down migration failed: %w", err)
		}
		log.Println("Successfully rolled back last migration")
		return nil
	}

	// Check for -all flag
	if args[0] == "-all" {
		log.Println("Rolling back ALL migrations...")
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("down migration failed: %w", err)
		}
		log.Println("Successfully rolled back all migrations")
		return nil
	}

	// Try to parse N as integer
	var n int
	if _, err := fmt.Sscanf(args[0], "%d", &n); err != nil {
		return fmt.Errorf("invalid number of migrations: %s", args[0])
	}

	log.Printf("Rolling back %d migrations...\n", n)
	if err := m.Steps(-n); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("down migration failed: %w", err)
	}
	log.Printf("Successfully rolled back %d migrations", n)
	return nil
}

// runVersion shows the current migration version
func runVersion(m *migrate.Migrate) error {
	version, dirty, err := m.Version()
	if err != nil {
		return fmt.Errorf("could not get version: %w", err)
	}

	fmt.Printf("Current version: %d\n", version)
	if dirty {
		fmt.Println("Database is DIRTY - manual intervention required")
	} else {
		fmt.Println("Database is clean")
	}
	return nil
}

// runForce forces the migration version
func runForce(m *migrate.Migrate, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("force requires a version number")
	}

	var version int
	if _, err := fmt.Sscanf(args[0], "%d", &version); err != nil {
		return fmt.Errorf("invalid version number: %s", args[0])
	}

	log.Printf("Forcing migration version to %d...\n", version)
	if err := m.Force(version); err != nil {
		return fmt.Errorf("force failed: %w", err)
	}

	log.Printf("Successfully forced version to %d", version)
	return nil
}
