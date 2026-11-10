package config

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/awnumar/memguard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testConfigContent = `
env: test
postgres:
  dsn: postgres://user:pass@localhost:5432/db
redis:
  addr: localhost:6379
  password: secret
  db: 1
monnify:
  api_key: test-api-key
  secret_key: test-secret-key
  contract_code: test-contract
  base_url: https://api.monnify.com
`
)

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(testConfigContent), 0644)
	require.NoError(t, err)

	// Set working directory to tmpDir so viper finds the config.
	origWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	cfg, err := New()
	require.NoError(t, err)
	defer cfg.Destroy()

	assert.Equal(t, "test", cfg.Env())
	assert.Equal(t, "postgres://user:pass@localhost:5432/db", cfg.Postgres().DSN())
	assert.Equal(t, "localhost:6379", cfg.Redis().Addr())
	assert.Equal(t, "secret", cfg.Redis().Password())
	assert.Equal(t, 1, cfg.Redis().DB())
	assert.Equal(t, "test-api-key", cfg.Monnify().APIKey())
	assert.Equal(t, "test-secret-key", cfg.Monnify().SecretKey())
	assert.Equal(t, "test-contract", cfg.Monnify().ContractCode())
	assert.Equal(t, "https://api.monnify.com", cfg.Monnify().BaseURL())
}

func TestFromEnv(t *testing.T) {
	// Set environment variables.
	os.Setenv("ENV", "production")
	os.Setenv("POSTGRES_DSN", "postgres://prod:pass@prodhost:5432/proddb")
	os.Setenv("REDIS_ADDR", "prodredis:6379")
	os.Setenv("REDIS_PASSWORD", "prodpass")
	os.Setenv("REDIS_DB", "2")
	os.Setenv("MONNIFY_API_KEY", "prod-api-key")
	os.Setenv("MONNIFY_SECRET_KEY", "prod-secret-key")
	os.Setenv("MONNIFY_CONTRACT_CODE", "prod-contract")
	os.Setenv("MONNIFY_BASE_URL", "https://api.prod.monnify.com")

	defer func() {
		os.Unsetenv("ENV")
		os.Unsetenv("POSTGRES_DSN")
		os.Unsetenv("REDIS_ADDR")
		os.Unsetenv("REDIS_PASSWORD")
		os.Unsetenv("REDIS_DB")
		os.Unsetenv("MONNIFY_API_KEY")
		os.Unsetenv("MONNIFY_SECRET_KEY")
		os.Unsetenv("MONNIFY_CONTRACT_CODE")
		os.Unsetenv("MONNIFY_BASE_URL")
	}()

	cfg, err := FromEnv()
	require.NoError(t, err)
	defer cfg.Destroy()

	assert.Equal(t, "production", cfg.Env())
	assert.Equal(t, "postgres://prod:pass@prodhost:5432/proddb", cfg.Postgres().DSN())
	assert.Equal(t, "prodredis:6379", cfg.Redis().Addr())
	assert.Equal(t, "prodpass", cfg.Redis().Password())
	assert.Equal(t, 2, cfg.Redis().DB())
	assert.Equal(t, "prod-api-key", cfg.Monnify().APIKey())
	assert.Equal(t, "prod-secret-key", cfg.Monnify().SecretKey())
	assert.Equal(t, "prod-contract", cfg.Monnify().ContractCode())
	assert.Equal(t, "https://api.prod.monnify.com", cfg.Monnify().BaseURL())
}

func TestReload(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(testConfigContent), 0644)
	require.NoError(t, err)

	origWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	cfg, err := New()
	require.NoError(t, err)
	defer cfg.Destroy()

	// Modify the config file.
	newContent := `
env: staging
postgres:
  dsn: postgres://staging:pass@staginghost:5432/stagingdb
redis:
  addr: stagingredis:6379
  password: stagingpass
  db: 3
monnify:
  api_key: staging-api-key
  secret_key: staging-secret-key
  contract_code: staging-contract
  base_url: https://api.staging.monnify.com
`
	err = os.WriteFile(configPath, []byte(newContent), 0644)
	require.NoError(t, err)

	// Reload.
	err = cfg.Reload()
	require.NoError(t, err)

	// Verify updated values.
	assert.Equal(t, "staging", cfg.Env())
	assert.Equal(t, "postgres://staging:pass@staginghost:5432/stagingdb", cfg.Postgres().DSN())
	assert.Equal(t, "stagingredis:6379", cfg.Redis().Addr())
	assert.Equal(t, "stagingpass", cfg.Redis().Password())
	assert.Equal(t, 3, cfg.Redis().DB())
	assert.Equal(t, "staging-api-key", cfg.Monnify().APIKey())
	assert.Equal(t, "staging-secret-key", cfg.Monnify().SecretKey())
	assert.Equal(t, "staging-contract", cfg.Monnify().ContractCode())
	assert.Equal(t, "https://api.staging.monnify.com", cfg.Monnify().BaseURL())
}

func TestWatch(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(testConfigContent), 0644)
	require.NoError(t, err)

	origWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	cfg, err := New()
	require.NoError(t, err)
	defer cfg.Destroy()

	// Channel to signal callback.
	callbackCalled := make(chan bool, 1)
	cfg.Watch(func(c Config) {
		// Verify updated values.
		assert.Equal(t, "staging", c.Env())
		assert.Equal(t, "postgres://staging:pass@staginghost:5432/stagingdb", c.Postgres().DSN())
		callbackCalled <- true
	})

	// Wait a bit for watcher to start.
	time.Sleep(100 * time.Millisecond)

	// Modify the file.
	newContent := `
env: staging
postgres:
  dsn: postgres://staging:pass@staginghost:5432/stagingdb
redis:
  addr: stagingredis:6379
  password: stagingpass
  db: 3
monnify:
  api_key: staging-api-key
  secret_key: staging-secret-key
  contract_code: staging-contract
  base_url: https://api.staging.monnify.com
`
	err = os.WriteFile(configPath, []byte(newContent), 0644)
	require.NoError(t, err)

	// Wait for callback or timeout.
	select {
	case <-callbackCalled:
		// Success.
	case <-time.After(2 * time.Second):
		t.Fatal("callback not called within timeout")
	}
}

func TestDestroy(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(testConfigContent), 0644)
	require.NoError(t, err)

	origWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	cfg, err := New()
	require.NoError(t, err)

	// Destroy.
	cfg.Destroy()

	// After destroy, getters should return zero values (empty strings, nil pointers).
	// The appConfig's mu is still there but the underlying configs are nil.
	// Check that the config is nil-safe.
	assert.Equal(t, "", cfg.Env())
	assert.Nil(t, cfg.Postgres())
	assert.Nil(t, cfg.Redis())
	assert.Nil(t, cfg.Monnify())

	// Reload should fail.
	err = cfg.Reload()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config has been destroyed")

	// Watch should be no-op.
	cfg.Watch(func(c Config) {})
}

func TestMissingRequiredKeys(t *testing.T) {
	// Create a config file missing required keys.
	content := `
env: test
postgres:
  dsn: postgres://...
# missing redis.addr
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err)

	origWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	_, err = New()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration key redis.addr is required")
}

func TestPostgresConnect(t *testing.T) {
	// We can't actually connect without a real DB, but we can test that the function
	// returns an error when DSN is empty or invalid.
	pg := &PostgresConfig{dsn: nil}
	_, err := pg.Connect(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "postgres DSN is empty")

	// Provide a fake DSN to ensure it attempts to parse and returns error.
	pg.dsn = memguard.NewBufferFromBytes([]byte("invalid dsn"))
	_, err = pg.Connect(context.Background())
	assert.Error(t, err)
	// The exact error may be "failed to parse config", we just check error not nil.
	assert.Contains(t, err.Error(), "failed to create postgres pool")
}

func TestRedisConnect(t *testing.T) {
	// Redis Connect just creates a client without error, we can test options.
	redisCfg := &RedisConfig{
		addr:     "localhost:6379",
		password: memguard.NewBufferFromBytes([]byte("pass")),
		db:       1,
	}
	client := redisCfg.Connect()
	assert.NotNil(t, client)
	// Options should match.
	opts := client.Options()
	assert.Equal(t, "localhost:6379", opts.Addr)
	assert.Equal(t, "pass", opts.Password)
	assert.Equal(t, 1, opts.DB)
}

// Additional test: Reload with environment variables priority.
func TestReloadWithEnvOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(testConfigContent), 0644)
	require.NoError(t, err)

	origWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	// Set environment variables to override some values.
	os.Setenv("POSTGRES_DSN", "postgres://env:pass@envhost:5432/envdb")
	os.Setenv("MONNIFY_API_KEY", "env-api-key")
	defer func() {
		os.Unsetenv("POSTGRES_DSN")
		os.Unsetenv("MONNIFY_API_KEY")
	}()

	cfg, err := New()
	require.NoError(t, err)
	defer cfg.Destroy()

	// Check that env vars take precedence.
	assert.Equal(t, "postgres://env:pass@envhost:5432/envdb", cfg.Postgres().DSN())
	assert.Equal(t, "env-api-key", cfg.Monnify().APIKey())
	// Other values should come from file.
	assert.Equal(t, "localhost:6379", cfg.Redis().Addr())
	assert.Equal(t, "test-secret-key", cfg.Monnify().SecretKey())
}

// Test that multiple watchers can be registered.
func TestMultipleWatchers(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(testConfigContent), 0644)
	require.NoError(t, err)

	origWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	cfg, err := New()
	require.NoError(t, err)
	defer cfg.Destroy()

	callCount := 0
	var mu sync.Mutex
	cb := func(c Config) {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	cfg.Watch(cb)
	cfg.Watch(cb)

	time.Sleep(100 * time.Millisecond)

	newContent := `
env: staging
postgres:
  dsn: postgres://staging:pass@staginghost:5432/stagingdb
redis:
  addr: stagingredis:6379
  password: stagingpass
  db: 3
monnify:
  api_key: staging-api-key
  secret_key: staging-secret-key
  contract_code: staging-contract
  base_url: https://api.staging.monnify.com
`
	err = os.WriteFile(configPath, []byte(newContent), 0644)
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)
	mu.Lock()
	assert.Equal(t, 2, callCount) // each callback should be called once
	mu.Unlock()
}
