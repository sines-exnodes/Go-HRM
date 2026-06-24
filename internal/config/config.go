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
	AppEnv         string
	Port           string
	GinMode        string
	SwaggerEnabled bool

	// CORSAllowedOrigins is a comma-separated list of origins permitted by
	// the CORS middleware. Empty is allowed only in development.
	CORSAllowedOrigins string

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

	// Brute-force protection on /auth/login. Defaults match the Python
	// repo: 5 bad passwords trigger a 15-minute account lockout.
	MaxFailedLoginAttempts int
	AccountLockoutMinutes  int

	// Remember-me flag on /auth/login extends the refresh-token TTL. Default
	// matches Python's REMEMBER_ME_REFRESH_TOKEN_EXPIRE_DAYS (30 days).
	RememberMeRefreshTTLDays int

	SuperAdminEmail    string
	SuperAdminPassword string
	SuperAdminName     string

	Storage StorageConfig

	// Attendance / office-presence settings (Phase 6). Read by the attendance
	// service to evaluate is_late, GPS proximity, and half-day flagging.
	// Defaults match the Python source (Asia/Ho_Chi_Minh, 09:00 lateness,
	// 18:00 early-leave cutoff, GPS disabled). When OFFICE_GPS_ENABLED=true,
	// OFFICE_LATITUDE / OFFICE_LONGITUDE / OFFICE_RADIUS_METERS define the
	// allowed check-in zone.
	CompanyTimezone         string
	LateThresholdHour       int
	LateThresholdMinute     int
	CheckoutThresholdHour   int
	CheckoutThresholdMinute int
	OfficeGPSEnabled        bool
	OfficeLatitude          float64
	OfficeLongitude         float64
	OfficeRadiusMeters      float64
	HalfDayHoursThreshold   float64

	// Email + Invite + Push (Phase 9). When SMTP_HOST is empty the
	// EmailService logs the would-be email and writes last_email_error
	// onto the invite row — invite creation still succeeds. Same shape
	// for FIREBASE_*: empty credentials path = PushClient is a no-op
	// logger (parity with Python's _get_firebase_app()).
	AppName                       string
	FrontendURL                   string
	InviteTokenExpireHours        int
	PasswordResetTokenExpireHours int
	SMTPHost                      string
	SMTPPort                      int
	SMTPUser                      string
	SMTPPassword                  string
	SMTPFromEmail                 string
	SMTPFromName                  string
	SMTPUseTLS                    bool
	FirebaseCredentialsPath       string
	FirebaseProjectID             string
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

		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", ""),

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

		MaxFailedLoginAttempts:   getEnvInt("MAX_FAILED_LOGIN_ATTEMPTS", 5),
		AccountLockoutMinutes:    getEnvInt("ACCOUNT_LOCKOUT_MINUTES", 15),
		RememberMeRefreshTTLDays: getEnvInt("REMEMBER_ME_REFRESH_TOKEN_EXPIRE_DAYS", 30),

		SuperAdminEmail:    getEnv("SUPER_ADMIN_EMAIL", ""),
		SuperAdminPassword: getEnv("SUPER_ADMIN_PASSWORD", ""),
		SuperAdminName:     getEnv("SUPER_ADMIN_NAME", "Super Admin"),

		CompanyTimezone:         getEnv("COMPANY_TIMEZONE", "Asia/Ho_Chi_Minh"),
		LateThresholdHour:       getEnvInt("LATE_THRESHOLD_HOUR", 9),
		LateThresholdMinute:     getEnvInt("LATE_THRESHOLD_MINUTE", 0),
		CheckoutThresholdHour:   getEnvInt("CHECKOUT_THRESHOLD_HOUR", 18),
		CheckoutThresholdMinute: getEnvInt("CHECKOUT_THRESHOLD_MINUTE", 0),
		OfficeGPSEnabled:        getEnvBool("OFFICE_GPS_ENABLED", false),
		OfficeLatitude:          getEnvFloat("OFFICE_LATITUDE", 0.0),
		OfficeLongitude:         getEnvFloat("OFFICE_LONGITUDE", 0.0),
		OfficeRadiusMeters:      getEnvFloat("OFFICE_RADIUS_METERS", 50.0),
		HalfDayHoursThreshold:   getEnvFloat("HALF_DAY_HOURS_THRESHOLD", 4.0),

		AppName:                       getEnv("APP_NAME", "Exnodes HRM"),
		FrontendURL:                   getEnv("FRONTEND_URL", "http://localhost:3000"),
		InviteTokenExpireHours:        getEnvInt("INVITE_TOKEN_EXPIRE_HOURS", 72),
		PasswordResetTokenExpireHours: getEnvInt("PASSWORD_RESET_TOKEN_EXPIRE_HOURS", 1),
		SMTPHost:                      getEnv("SMTP_HOST", ""),
		SMTPPort:                      getEnvInt("SMTP_PORT", 587),
		SMTPUser:                      getEnv("SMTP_USER", ""),
		SMTPPassword:                  getEnv("SMTP_PASSWORD", ""),
		SMTPFromEmail:                 getEnv("SMTP_FROM_EMAIL", ""),
		SMTPFromName:                  getEnv("SMTP_FROM_NAME", "Exnodes HRM"),
		SMTPUseTLS:                    getEnvBool("SMTP_USE_TLS", true),
		FirebaseCredentialsPath:       getEnv("FIREBASE_CREDENTIALS_PATH", ""),
		FirebaseProjectID:             getEnv("FIREBASE_PROJECT_ID", ""),
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

func getEnvFloat(key string, fallback float64) float64 {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
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
