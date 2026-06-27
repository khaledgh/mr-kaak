// Package config loads and validates application configuration from the
// environment. Configuration is parsed once at startup into an immutable
// struct; nothing else in the app reads os.Getenv directly.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment describes the deployment environment.
type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
)

// Config is the fully-validated application configuration.
type Config struct {
	App   App
	HTTP  HTTP
	Log   Log
	DB    DB
	Redis Redis
	Meili Meili
	JWT   JWT
	I18n  I18n
	VAPID VAPID
	Media Media
}

type Media struct {
	UploadDir     string   // local directory to write files, default ./uploads
	PublicBaseURL string   // absolute base URL prefix, e.g. http://127.0.0.1:8080
	MaxBytes      int64    // per-file hard limit (bytes)
	AllowedMIME   []string // e.g. ["image/jpeg", "image/png", "image/webp"]
}

type App struct {
	Env             Environment
	Name            string
	ShutdownTimeout time.Duration
}

func (a App) IsProduction() bool { return a.Env == EnvProduction }

type HTTP struct {
	Port           int
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	AllowedOrigins []string
}

type Log struct {
	Level  string
	Format string // console | json
}

type DB struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	Params          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DSN builds the MySQL data source name for GORM.
func (d DB) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.Params)
}

type Redis struct {
	Addr     string
	Password string
	DB       int // primary DB for Asynq/pubsub
	CacheDB  int // dedicated DB index for cache-aside (so FLUSHDB is safe)
}

type Meili struct {
	Host      string
	MasterKey string
}

type JWT struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type I18n struct {
	DefaultLocale string
	Currency      string
}

// VAPID holds Web Push keys (plan §8). Empty keys disable push (no-op).
type VAPID struct {
	PublicKey  string
	PrivateKey string
	Subject    string // mailto: or https: contact, per Web Push spec
}

// Load reads configuration from the environment and validates it.
// It returns a descriptive error listing every problem found, so a
// misconfigured deploy fails fast at startup instead of at first request.
func Load() (*Config, error) {
	v := newReader()

	cfg := &Config{
		App: App{
			Env:             Environment(v.str("APP_ENV", "development")),
			Name:            v.str("APP_NAME", "restaurant-api"),
			ShutdownTimeout: v.dur("SHUTDOWN_TIMEOUT", 20*time.Second),
		},
		HTTP: HTTP{
			Port:           v.intv("HTTP_PORT", 8080),
			ReadTimeout:    v.dur("HTTP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:   v.dur("HTTP_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:    v.dur("HTTP_IDLE_TIMEOUT", 60*time.Second),
			AllowedOrigins: v.csv("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
		},
		Log: Log{
			Level:  v.str("LOG_LEVEL", "info"),
			Format: v.str("LOG_FORMAT", "json"),
		},
		DB: DB{
			Host:            v.str("DB_HOST", "127.0.0.1"),
			Port:            v.intv("DB_PORT", 3306),
			User:            v.str("DB_USER", "restaurant"),
			Password:        v.str("DB_PASSWORD", ""),
			Name:            v.str("DB_NAME", "restaurant"),
			Params:          v.str("DB_PARAMS", "charset=utf8mb4&parseTime=true&loc=UTC"),
			MaxOpenConns:    v.intv("DB_MAX_OPEN_CONNS", 50),
			MaxIdleConns:    v.intv("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: v.dur("DB_CONN_MAX_LIFETIME", 30*time.Minute),
			ConnMaxIdleTime: v.dur("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
		Redis: Redis{
			Addr:     v.str("REDIS_ADDR", "127.0.0.1:6379"),
			Password: v.str("REDIS_PASSWORD", ""),
			DB:       v.intv("REDIS_DB", 0),
			CacheDB:  v.intv("REDIS_CACHE_DB", 1),
		},
		Meili: Meili{
			Host:      v.str("MEILI_HOST", "http://127.0.0.1:7700"),
			MasterKey: v.str("MEILI_MASTER_KEY", ""),
		},
		JWT: JWT{
			AccessSecret:  v.str("JWT_ACCESS_SECRET", ""),
			RefreshSecret: v.str("JWT_REFRESH_SECRET", ""),
			AccessTTL:     v.dur("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL:    v.dur("JWT_REFRESH_TTL", 720*time.Hour),
		},
		I18n: I18n{
			DefaultLocale: v.str("DEFAULT_LOCALE", "en"),
			Currency:      v.str("CURRENCY", "CAD"),
		},
		VAPID: VAPID{
			PublicKey:  v.str("VAPID_PUBLIC_KEY", ""),
			PrivateKey: v.str("VAPID_PRIVATE_KEY", ""),
			Subject:    v.str("VAPID_SUBJECT", "mailto:admin@example.com"),
		},
		Media: Media{
			UploadDir:     v.str("UPLOAD_DIR", "./uploads"),
			PublicBaseURL: v.str("PUBLIC_BASE_URL", "http://127.0.0.1:8080"),
			MaxBytes:      int64(v.intv("MEDIA_MAX_BYTES", 2*1024*1024)),
			AllowedMIME:   v.csv("MEDIA_ALLOWED_MIME", []string{"image/jpeg", "image/png", "image/webp"}),
		},
	}

	if errs := append(v.errors, cfg.validate()...); len(errs) > 0 {
		return nil, fmt.Errorf("invalid configuration:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return cfg, nil
}

// validate enforces invariants that depend on the loaded values.
func (c *Config) validate() []string {
	var errs []string

	switch c.App.Env {
	case EnvDevelopment, EnvStaging, EnvProduction:
	default:
		errs = append(errs, fmt.Sprintf("APP_ENV %q must be development|staging|production", c.App.Env))
	}
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		errs = append(errs, fmt.Sprintf("HTTP_PORT %d out of range", c.HTTP.Port))
	}
	if c.DB.Name == "" {
		errs = append(errs, "DB_NAME is required")
	}

	// Production must not run on dev defaults / empty secrets.
	if c.App.IsProduction() {
		if isWeakSecret(c.JWT.AccessSecret) {
			errs = append(errs, "JWT_ACCESS_SECRET must be a strong secret in production")
		}
		if isWeakSecret(c.JWT.RefreshSecret) {
			errs = append(errs, "JWT_REFRESH_SECRET must be a strong secret in production")
		}
		if c.Meili.MasterKey == "" {
			errs = append(errs, "MEILI_MASTER_KEY is required in production")
		}
		for _, o := range c.HTTP.AllowedOrigins {
			if o == "*" {
				errs = append(errs, "CORS_ALLOWED_ORIGINS must not contain '*' in production")
			}
		}
	}
	return errs
}

func isWeakSecret(s string) bool {
	return len(s) < 24 || strings.Contains(strings.ToLower(s), "change") || strings.Contains(strings.ToLower(s), "dev-")
}

// reader accumulates parse errors instead of panicking on the first one.
type reader struct{ errors []string }

func newReader() *reader { return &reader{} }

func (r *reader) str(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func (r *reader) intv(key string, def int) int {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		r.errors = append(r.errors, fmt.Sprintf("%s %q is not an integer", key, raw))
		return def
	}
	return n
}

func (r *reader) dur(key string, def time.Duration) time.Duration {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		r.errors = append(r.errors, fmt.Sprintf("%s %q is not a duration (e.g. 10s, 5m)", key, raw))
		return def
	}
	return d
}

func (r *reader) csv(key string, def []string) []string {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return def
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
