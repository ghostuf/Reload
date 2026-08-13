package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL     string
	Port            string
	JWTAccessSecret string
	CookieDomain    string
	CookieSecure    bool
	// FrontendOrigin is the exact origin allowed to make credentialed requests.
	// Cookie auth rules out a wildcard, so this cannot be "*".
	FrontendOrigin string
}

// ErrMissingJWTSecret is returned when JWT_ACCESS_SECRET is unset. Falling back
// to a default would let every deployment sign tokens the others accept.
var ErrMissingJWTSecret = errors.New("JWT_ACCESS_SECRET must be set")

func Load() Config {
	_ = godotenv.Load()

	return Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://reload:reload@localhost:5432/reloaddb?sslmode=disable"),
		Port:            getEnv("PORT", "8080"),
		JWTAccessSecret: getEnv("JWT_ACCESS_SECRET", ""),
		CookieDomain:    getEnv("COOKIE_DOMAIN", ""),
		CookieSecure:    getEnv("COOKIE_SECURE", "false") == "true",
		FrontendOrigin:  getEnv("FRONTEND_ORIGIN", "http://localhost:3000"),
	}
}

// Validate reports configuration that would make the server unsafe or unable to
// serve auth, so startup can fail loudly instead of at the first login.
func (c Config) Validate() error {
	if c.JWTAccessSecret == "" {
		return ErrMissingJWTSecret
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
