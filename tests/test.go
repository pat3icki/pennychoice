package testdb

import (
	"path/filepath"

	"github.com/pat3icki/pkgt/envt" // Replace with your actual module path
)

const (
	REMOTE_NAME     = "remote_db.env"
	LOCAL_NAME      = "local_db.env"
	REMOTE_DB_ROUTE = true
)

func LoadTestEnv() (map[string]string, string) {
	// If we are not forced to remote, try loading the local env first
	if !REMOTE_DB_ROUTE {
		envMap, err := envt.Map(&envt.Config{
			EnvPath:  filepath.Join("tests", LOCAL_NAME),
			PanicErr: false, // Don't panic if local doesn't exist, just fallback
		})
		if err == nil {
			return envMap, "local"
		}
	}

	// Otherwise, load remote (PanicErr: true ensures it fails loudly if missing, matching original behavior)
	envMap, err := envt.Map(&envt.Config{
		EnvPath:  filepath.Join("tests", REMOTE_NAME),
		PanicErr: true,
	})
	if err != nil {
		// This block is technically unreachable if PanicErr is true,
		// but kept here for explicit clarity or if you change PanicErr to false.
		panic("failed to load remote env: " + err.Error())
	}

	return envMap, "remote"
}
