package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	var (
		user     string
		password string
		dbName   string
		portMap  string
	)

	flag.StringVar(&user, "POSTGRES_USER", "myapp", "PostgreSQL user")
	flag.StringVar(&password, "POSTGRES_PASSWORD", "secret", "PostgreSQL password")
	flag.StringVar(&dbName, "POSTGRES_DB", "myapp", "PostgreSQL database name")
	flag.StringVar(&portMap, "PORT_MAP", "5432:5432", "Port mapping (host:container)")
	flag.Parse()

	// Extract host port from PORT_MAP (left side)
	parts := strings.Split(portMap, ":")
	if len(parts) != 2 {
		log.Fatalf("invalid PORT_MAP format: %s (expected host:container)", portMap)
	}
	hostPort := parts[0]

	// Build the database URL
	url := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", user, password, hostPort, dbName)

	// Locate the project root (same logic as in test.go)
	root := findProjectRoot()

	// Destination: tests/local_db.env
	envDir := filepath.Join(root, "tests")
	if err := os.MkdirAll(envDir, 0755); err != nil {
		log.Fatalf("failed to create tests directory: %v", err)
	}
	envPath := filepath.Join(envDir, "local_db.env")

	// Write the variable to the file (overwrites if exists)
	content := fmt.Sprintf("TEST_DATABASE_URL=%s\n", url)
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		log.Fatalf("failed to write env file: %v", err)
	}

	log.Printf("Successfully wrote %s with TEST_DATABASE_URL", envPath)
}

// findProjectRoot climbs up until it finds go.mod (same as in test.go)
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
