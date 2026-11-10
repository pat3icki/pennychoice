package db

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pat3icki/pennychoice/internal/db/cache"
	"github.com/pat3icki/pennychoice/internal/db/sqlc"
)

// Config holds database connection configuration
type Config struct {
	// Connection URI takes precedence over individual parameters if provided
	URI string // PostgreSQL connection URI (e.g., postgres://user:pass@host:5432/db?sslmode=require)

	// Individual connection parameters (used if URI is empty)
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string // "disable", "require", "verify-ca", "verify-full"

	// Certificate configuration - either file paths or raw data
	// Priority: Data fields > File path fields
	SSLCertPath   string // Path to client certificate (optional)
	SSLKeyPath    string // Path to client key (optional)
	SSLRootCAPath string // Path to CA certificate (required for verify-ca/verify-full)

	SSLCertData   []byte // Raw PEM data for client certificate (optional)
	SSLKeyData    []byte // Raw PEM data for client key (optional)
	SSLRootCAData []byte // Raw PEM data for CA certificate (optional)

	// Connection pool settings
	MaxConnections    int32
	MinConnections    int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

// NewClient creates a new database client with connection pooling
func NewClient(ctx context.Context, cfg Config, cacheClient cache.Cache) (*Client, error) {
	var poolConfig *pgxpool.Config
	var err error

	// Extract host for ServerName if using URI
	var serverName string

	// Parse connection configuration
	if cfg.URI != "" {
		// Use URI connection string
		poolConfig, err = pgxpool.ParseConfig(cfg.URI)
		if err != nil {
			return nil, fmt.Errorf("failed to parse connection URI: %w", err)
		}

		// Extract host from URI for ServerName
		parsedURL, err := url.Parse(cfg.URI)
		if err == nil {
			serverName = parsedURL.Hostname()
			// If port is specified, we need to strip it
			if strings.Contains(serverName, ":") {
				serverName = strings.Split(serverName, ":")[0]
			}
		}

		// Override SSL/TLS configuration if certificate data is provided
		if cfg.SSLRootCAData != nil || cfg.SSLCertData != nil || cfg.SSLRootCAPath != "" || cfg.SSLCertPath != "" {
			tlsConfig, err := configureTLSFromURI(cfg, serverName)
			if err != nil {
				return nil, fmt.Errorf("failed to configure TLS for URI: %w", err)
			}
			poolConfig.ConnConfig.TLSConfig = tlsConfig
		} else {
			// Even without custom certs, ensure ServerName is set
			if poolConfig.ConnConfig.TLSConfig != nil && poolConfig.ConnConfig.TLSConfig.ServerName == "" {
				if serverName != "" {
					poolConfig.ConnConfig.TLSConfig.ServerName = serverName
				} else {
					// Fallback: use the host from the config
					poolConfig.ConnConfig.TLSConfig.ServerName = cfg.Host
				}
			}
		}
	} else {
		// Build connection string from individual parameters
		connString := buildConnectionString(cfg)
		poolConfig, err = pgxpool.ParseConfig(connString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse connection config: %w", err)
		}

		// Configure TLS/SSL based on sslmode
		if cfg.SSLMode != "disable" {
			tlsConfig, err := configureTLS(cfg)
			if err != nil {
				return nil, fmt.Errorf("failed to configure TLS: %w", err)
			}
			poolConfig.ConnConfig.TLSConfig = tlsConfig
		}
	}

	// Ensure ServerName is set for TLS connections
	if poolConfig.ConnConfig.TLSConfig != nil {
		if poolConfig.ConnConfig.TLSConfig.ServerName == "" {
			// Try to get ServerName from host
			if cfg.Host != "" {
				poolConfig.ConnConfig.TLSConfig.ServerName = cfg.Host
			} else if serverName != "" {
				poolConfig.ConnConfig.TLSConfig.ServerName = serverName
			} else {
				// For Aiven and similar services, we need the full hostname
				// Try to extract from the connection string
				if poolConfig.ConnConfig.Host != "" {
					poolConfig.ConnConfig.TLSConfig.ServerName = poolConfig.ConnConfig.Host
				}
			}
		}
	}

	// Configure connection pool
	if cfg.MaxConnections > 0 {
		poolConfig.MaxConns = cfg.MaxConnections
	}
	if cfg.MinConnections > 0 {
		poolConfig.MinConns = cfg.MinConnections
	}
	if cfg.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	}
	if cfg.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	}
	if cfg.HealthCheckPeriod > 0 {
		poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod
	}

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Ping to verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Client{
		_pool: pool,
		Query: sqlc.New(pool),
		Cache: cacheClient,
	}, nil
}

// buildConnectionString constructs the PostgreSQL connection string
func buildConnectionString(cfg Config) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
	)
}

// getCertificateData returns certificate data from either raw data or file path
// Priority: Data > File path
func getCertificateData(data []byte, filePath string) ([]byte, error) {
	if len(data) > 0 {
		return data, nil
	}

	if filePath != "" {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read certificate file %s: %w", filePath, err)
		}
		return content, nil
	}

	return nil, nil
}

// configureTLS sets up TLS configuration for the connection
func configureTLS(cfg Config) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Set ServerName for host verification
	if cfg.Host != "" {
		tlsConfig.ServerName = cfg.Host
	}

	// Handle CA certificate
	if cfg.SSLMode == "verify-ca" || cfg.SSLMode == "verify-full" {
		caData, err := getCertificateData(cfg.SSLRootCAData, cfg.SSLRootCAPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get CA certificate: %w", err)
		}

		if caData == nil || len(caData) == 0 {
			return nil, fmt.Errorf("root CA certificate required for sslmode %s", cfg.SSLMode)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caData) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Handle client certificates (mutual TLS)
	certData, err := getCertificateData(cfg.SSLCertData, cfg.SSLCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get client certificate: %w", err)
	}

	keyData, err := getCertificateData(cfg.SSLKeyData, cfg.SSLKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get client key: %w", err)
	}

	if certData != nil && len(certData) > 0 && keyData != nil && len(keyData) > 0 {
		cert, err := tls.X509KeyPair(certData, keyData)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

// configureTLSFromURI configures TLS for URI-based connections
func configureTLSFromURI(cfg Config, serverName string) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Set ServerName for host verification
	if serverName != "" {
		tlsConfig.ServerName = serverName
	} else if cfg.Host != "" {
		tlsConfig.ServerName = cfg.Host
	}

	// Handle CA certificate
	caData, err := getCertificateData(cfg.SSLRootCAData, cfg.SSLRootCAPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get CA certificate: %w", err)
	}

	if caData != nil && len(caData) > 0 {
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caData) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Handle client certificates (mutual TLS)
	certData, err := getCertificateData(cfg.SSLCertData, cfg.SSLCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get client certificate: %w", err)
	}

	keyData, err := getCertificateData(cfg.SSLKeyData, cfg.SSLKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get client key: %w", err)
	}

	if certData != nil && len(certData) > 0 && keyData != nil && len(keyData) > 0 {
		cert, err := tls.X509KeyPair(certData, keyData)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// If no CA certificate was provided but sslmode requires it,
	// we should still have a valid TLS config
	if tlsConfig.RootCAs == nil {
		// Use system CA pool as fallback
		systemPool, err := x509.SystemCertPool()
		if err == nil && systemPool != nil {
			tlsConfig.RootCAs = systemPool
		}
	}

	return tlsConfig, nil
}

// NewClientFromEnv creates a client using environment variables
func NewClientFromEnv(ctx context.Context, cacheClient cache.Cache) (*Client, error) {
	cfg := Config{
		// URI takes precedence if provided
		URI:            os.Getenv("DATABASE_URL"),
		Host:           getEnvOrDefault("DB_HOST", "localhost"),
		Port:           getEnvIntOrDefault("DB_PORT", 5432),
		User:           getEnvOrDefault("DB_USER", "postgres"),
		Password:       getEnvOrDefault("DB_PASSWORD", ""),
		Database:       getEnvOrDefault("DB_NAME", "postgres"),
		SSLMode:        getEnvOrDefault("DB_SSL_MODE", "disable"),
		SSLCertPath:    getEnvOrDefault("DB_SSL_CERT_PATH", ""),
		SSLKeyPath:     getEnvOrDefault("DB_SSL_KEY_PATH", ""),
		SSLRootCAPath:  getEnvOrDefault("DB_SSL_CA_PATH", ""),
		MaxConnections: int32(getEnvIntOrDefault("DB_MAX_CONNECTIONS", 10)),
		MinConnections: int32(getEnvIntOrDefault("DB_MIN_CONNECTIONS", 2)),
	}
	return NewClient(ctx, cfg, cacheClient)
}

// getEnvOrDefault returns environment variable or default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault returns environment variable as int or default value
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
