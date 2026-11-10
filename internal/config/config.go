package config

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/awnumar/memguard"
	"github.com/fsnotify/fsnotify"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

// Config defines the interface for accessing application configuration.
type Config interface {
	Env() string
	Postgres() *PostgresConfig
	Redis() *RedisConfig
	Monnify() *MonnifyConfig
	Reload() error         // re‑read config file and update values
	Destroy()              // wipe all secrets
	Watch(cb func(Config)) // watch for file changes and invoke callback
}

// PostgresConfig holds PostgreSQL connection settings.
type PostgresConfig struct {
	dsn *memguard.LockedBuffer
}

// DSN returns the PostgreSQL connection string.
func (p *PostgresConfig) DSN() string {
	return getSecretFromBuffer(p.dsn)
}

// Connect creates a new pgx/v5 connection pool.
// The caller is responsible for calling Close() on the returned pool.
func (p *PostgresConfig) Connect(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := p.DSN()
	if dsn == "" {
		return nil, fmt.Errorf("postgres DSN is empty")
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}
	return pool, nil
}

// Destroy wipes sensitive postgres data from memory.
func (p *PostgresConfig) Destroy() {
	if p.dsn != nil {
		p.dsn.Destroy()
	}
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	addr     string
	password *memguard.LockedBuffer
	db       int
}

// Addr returns the Redis server address.
func (r *RedisConfig) Addr() string { return r.addr }

// Password returns the Redis password.
func (r *RedisConfig) Password() string {
	return getSecretFromBuffer(r.password)
}

// DB returns the Redis database number.
func (r *RedisConfig) DB() int { return r.db }

// Connect creates a new Redis client.
// The caller is responsible for calling Close() on the returned client.
func (r *RedisConfig) Connect() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     r.addr,
		Password: r.Password(),
		DB:       r.db,
	})
}

// Destroy wipes sensitive redis data from memory.
func (r *RedisConfig) Destroy() {
	if r.password != nil {
		r.password.Destroy()
	}
}

// MonnifyConfig holds Monnify API credentials.
type MonnifyConfig struct {
	apiKey       *memguard.LockedBuffer
	secretKey    *memguard.LockedBuffer
	contractCode *memguard.LockedBuffer
	baseURL      string
}

// APIKey returns the Monnify API key.
func (m *MonnifyConfig) APIKey() string {
	return getSecretFromBuffer(m.apiKey)
}

// SecretKey returns the Monnify secret key.
func (m *MonnifyConfig) SecretKey() string {
	return getSecretFromBuffer(m.secretKey)
}

// ContractCode returns the Monnify contract code.
func (m *MonnifyConfig) ContractCode() string {
	return getSecretFromBuffer(m.contractCode)
}

// BaseURL returns the Monnify base URL.
func (m *MonnifyConfig) BaseURL() string { return m.baseURL }

// Destroy wipes sensitive monnify data from memory.
func (m *MonnifyConfig) Destroy() {
	if m.apiKey != nil {
		m.apiKey.Destroy()
	}
	if m.secretKey != nil {
		m.secretKey.Destroy()
	}
	if m.contractCode != nil {
		m.contractCode.Destroy()
	}
}

type configData struct {
	env      string
	postgres *PostgresConfig
	redis    *RedisConfig
	monnify  *MonnifyConfig
}

type appConfig struct {
	mu         sync.RWMutex
	env        string
	postgres   *PostgresConfig
	redis      *RedisConfig
	monnify    *MonnifyConfig
	v          *viper.Viper
	configFile string

	watchMu   sync.Mutex
	watcher   *fsnotify.Watcher
	callbacks []func(Config)
	done      chan struct{}
	wg        sync.WaitGroup
	destroyed atomic.Bool
}

// New creates a new Config instance. It loads configuration from the following sources
// in order of precedence: explicit config file (config.yaml), environment variables.
// Sensitive variables are locked into protected memory using memguard.
func New() (Config, error) {
	v := viper.New()
	v.SetDefault("env", "development")
	v.SetDefault("redis.db", 0)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./internal/config")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	data, err := buildConfig(v)
	if err != nil {
		return nil, err
	}

	return &appConfig{
		env:        data.env,
		postgres:   data.postgres,
		redis:      data.redis,
		monnify:    data.monnify,
		v:          v,
		configFile: v.ConfigFileUsed(),
		done:       make(chan struct{}),
	}, nil
}

// FromEnv creates a Config directly from environment variables.
func FromEnv() (Config, error) {
	v := viper.New()
	v.SetDefault("env", "development")
	v.SetDefault("redis.db", 0)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	data, err := buildConfig(v)
	if err != nil {
		return nil, err
	}

	return &appConfig{
		env:      data.env,
		postgres: data.postgres,
		redis:    data.redis,
		monnify:  data.monnify,
		v:        v,
		done:     make(chan struct{}),
	}, nil
}

func buildConfig(v *viper.Viper) (*configData, error) {
	getRequired := func(key string) (string, error) {
		if !v.IsSet(key) {
			return "", fmt.Errorf("configuration key %s is required but not set", key)
		}
		return v.GetString(key), nil
	}

	pgDSN, err := getRequired("postgres.dsn")
	if err != nil {
		return nil, err
	}

	redisAddr, err := getRequired("redis.addr")
	if err != nil {
		return nil, err
	}

	monnifyAPIKey, err := getRequired("monnify.api_key")
	if err != nil {
		return nil, err
	}
	monnifySecret, err := getRequired("monnify.secret_key")
	if err != nil {
		return nil, err
	}
	monnifyContract, err := getRequired("monnify.contract_code")
	if err != nil {
		return nil, err
	}

	return &configData{
		env: v.GetString("env"),
		postgres: &PostgresConfig{
			dsn: memguard.NewBufferFromBytes([]byte(pgDSN)),
		},
		redis: &RedisConfig{
			addr:     redisAddr,
			password: newBufferFromString(v.GetString("redis.password")),
			db:       v.GetInt("redis.db"),
		},
		monnify: &MonnifyConfig{
			apiKey:       memguard.NewBufferFromBytes([]byte(monnifyAPIKey)),
			secretKey:    memguard.NewBufferFromBytes([]byte(monnifySecret)),
			contractCode: memguard.NewBufferFromBytes([]byte(monnifyContract)),
			baseURL:      v.GetString("monnify.base_url"),
		},
	}, nil
}

func newBufferFromString(s string) *memguard.LockedBuffer {
	if s == "" {
		return nil
	}
	return memguard.NewBufferFromBytes([]byte(s))
}

func getSecretFromBuffer(buf *memguard.LockedBuffer) string {
	if buf == nil {
		return ""
	}
	return buf.String()
}

// Env returns the application environment.
func (c *appConfig) Env() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.env
}

// Postgres returns the current PostgreSQL configuration.
func (c *appConfig) Postgres() *PostgresConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.postgres
}

// Redis returns the current Redis configuration.
func (c *appConfig) Redis() *RedisConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.redis
}

// Monnify returns the current Monnify configuration.
func (c *appConfig) Monnify() *MonnifyConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.monnify
}

// Reload re-reads the configuration file and updates all values.
// Environment variables are read live and do not require reloading.
func (c *appConfig) Reload() error {
	if c.destroyed.Load() {
		return fmt.Errorf("config has been destroyed")
	}

	c.mu.Lock()

	if c.v != nil && c.configFile != "" {
		if err := c.v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				c.mu.Unlock()
				return fmt.Errorf("failed to reload config file: %w", err)
			}
		}
	}

	data, err := buildConfig(c.v)
	if err != nil {
		c.mu.Unlock()
		return fmt.Errorf("failed to build config during reload: %w", err)
	}

	oldPostgres := c.postgres
	oldRedis := c.redis
	oldMonnify := c.monnify

	c.env = data.env
	c.postgres = data.postgres
	c.redis = data.redis
	c.monnify = data.monnify
	c.configFile = c.v.ConfigFileUsed()

	c.mu.Unlock()

	// Destroy old configs outside the lock so existing readers are not blocked.
	if oldPostgres != nil {
		oldPostgres.Destroy()
	}
	if oldRedis != nil {
		oldRedis.Destroy()
	}
	if oldMonnify != nil {
		oldMonnify.Destroy()
	}

	return nil
}

// Destroy wipes all sensitive configuration from memory and stops file watchers.
// It is safe to call multiple times.
func (c *appConfig) Destroy() {
	if !c.destroyed.CompareAndSwap(false, true) {
		return
	}

	c.mu.Lock()
	if c.postgres != nil {
		c.postgres.Destroy()
		c.postgres = nil
	}
	if c.redis != nil {
		c.redis.Destroy()
		c.redis = nil
	}
	if c.monnify != nil {
		c.monnify.Destroy()
		c.monnify = nil
	}
	c.mu.Unlock()

	c.watchMu.Lock()
	if c.watcher != nil {
		c.watcher.Close()
		c.watcher = nil
	}
	c.watchMu.Unlock()

	close(c.done)
	c.wg.Wait()
}

// Watch registers a callback to be invoked when the config file changes on disk.
// The first call to Watch starts an fsnotify watcher on the config file's directory.
// Callbacks are invoked after the config is successfully reloaded.
func (c *appConfig) Watch(cb func(Config)) {
	if c.destroyed.Load() {
		return
	}

	c.watchMu.Lock()
	defer c.watchMu.Unlock()

	if c.destroyed.Load() {
		return
	}

	c.callbacks = append(c.callbacks, cb)

	if c.watcher != nil || c.configFile == "" {
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		// In production, log this error.
		return
	}
	c.watcher = watcher

	dir := filepath.Dir(c.configFile)
	if err := watcher.Add(dir); err != nil {
		watcher.Close()
		c.watcher = nil
		// In production, log this error.
		return
	}

	c.wg.Add(1)
	go c.watchLoop(c.configFile)
}

func (c *appConfig) watchLoop(configFile string) {
	defer c.wg.Done()

	baseFile := filepath.Base(configFile)
	var debounceTimer *time.Timer
	var debounceMu sync.Mutex

	for {
		select {
		case <-c.done:
			return
		case event, ok := <-c.watcher.Events:
			if !ok {
				return
			}
			if filepath.Base(event.Name) != baseFile {
				continue
			}
			if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) && !event.Has(fsnotify.Rename) {
				continue
			}

			debounceMu.Lock()
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(200*time.Millisecond, func() {
				debounceMu.Lock()
				defer debounceMu.Unlock()
				c.reloadAndNotify()
			})
			debounceMu.Unlock()

		case err, ok := <-c.watcher.Errors:
			if !ok {
				return
			}
			// In production, log the error.
			_ = err
		}
	}
}

// reloadAndNotify performs a reload and invokes all registered callbacks
// if the reload succeeds. It is safe to call from the watch loop.
func (c *appConfig) reloadAndNotify() {
	if c.destroyed.Load() {
		return
	}
	if err := c.Reload(); err != nil {
		// In production, log the error.
		return
	}
	c.watchMu.Lock()
	// Copy callbacks to avoid holding the lock while invoking them.
	cbs := make([]func(Config), len(c.callbacks))
	copy(cbs, c.callbacks)
	c.watchMu.Unlock()

	for _, cb := range cbs {
		cb(c)
	}
}
