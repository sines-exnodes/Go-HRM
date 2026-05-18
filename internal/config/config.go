package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all environment-driven settings for the API server.
type Config struct {
	AppEnv          string
	Port            string
	GinMode         string
	SwaggerEnabled  bool

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	DBUrl      string // explicit override; takes precedence over DB_* if non-empty

	MigrationsDir string

	JWTSecret           string
	JWTAccessTTLMinutes int
	JWTRefreshTTLHours  int
	JWTRefreshTTLDays   int

	SuperAdminEmail    string
	SuperAdminPassword string
	SuperAdminName     string

	SupabaseURL            string
	SupabaseServiceRoleKey string
	SupabaseBucket         string
	SupabaseS3Endpoint     string
	SupabaseS3Region       string
	SupabaseS3AccessKey    string
	SupabaseS3SecretKey    string

	Storage StorageConfig
}

// Load reads the .env file (if present) and returns the populated Config.
// Missing required values cause os.Exit via log.Fatal to avoid booting in a
// half-configured state.
func Load() *Config {
	_ = godotenv.Load() // .env is optional in CI/production

	cfg := &Config{
		AppEnv:         getEnv("APP_ENV", "development"),
		Port:           getEnv("PORT", "8080"),
		GinMode:        getEnv("GIN_MODE", "debug"),
		SwaggerEnabled: getEnvBool("SWAGGER_ENABLED", true),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "exnodes_hrm"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		DBUrl:      getEnv("DATABASE_URL", ""),

		MigrationsDir: getEnv("MIGRATIONS_DIR", "migrations"),

		JWTSecret:           getEnv("JWT_SECRET_KEY", ""),
		JWTAccessTTLMinutes: getEnvInt("JWT_ACCESS_TTL_MINUTES", 60),
		JWTRefreshTTLHours:  getEnvInt("JWT_REFRESH_TTL_HOURS", 720),
		JWTRefreshTTLDays:   getEnvInt("JWT_REFRESH_TTL_DAYS", 14),

		SuperAdminEmail:    getEnv("SUPER_ADMIN_EMAIL", ""),
		SuperAdminPassword: getEnv("SUPER_ADMIN_PASSWORD", ""),
		SuperAdminName:     getEnv("SUPER_ADMIN_NAME", "Super Admin"),

		SupabaseURL:            getEnv("SUPABASE_URL", ""),
		SupabaseServiceRoleKey: getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		SupabaseBucket:         getEnv("SUPABASE_BUCKET", ""),
		SupabaseS3Endpoint:     getEnv("SUPABASE_S3_ENDPOINT", ""),
		SupabaseS3Region:       getEnv("SUPABASE_S3_REGION", "ap-southeast-1"),
		SupabaseS3AccessKey:    getEnv("SUPABASE_S3_ACCESS_KEY", ""),
		SupabaseS3SecretKey:    getEnv("SUPABASE_S3_SECRET_KEY", ""),
	}

	if strings.TrimSpace(cfg.JWTSecret) == "" {
		log.Fatal("config: JWT_SECRET_KEY must be set")
	}

	// Storage is optional at boot; the upload module surfaces a clear error
	// if it is used while unconfigured. Only attach it when fully provided.
	if storage, err := LoadStorageConfig(); err == nil {
		cfg.Storage = storage
	}

	return cfg
}

// DSN returns a libpq-style DSN suitable for GORM's postgres driver.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// DatabaseURL returns a postgres:// URL suitable for golang-migrate. If the
// DATABASE_URL env var was set it is returned verbatim; otherwise one is
// composed from the DB_* parts.
func (c *Config) DatabaseURL() string {
	if strings.TrimSpace(c.DBUrl) != "" {
		return c.DBUrl
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		switch strings.ToLower(v) {
		case "1", "true", "yes", "y", "on":
			return true
		case "0", "false", "no", "n", "off":
			return false
		}
	}
	return fallback
}
