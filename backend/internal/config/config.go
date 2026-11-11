package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL     string
	Port            string
	JWTAccessSecret string
	CookieDomain    string
	CookieSecure    bool
	// CookieSameSite is "lax", "strict", or "none" (case-insensitive).
	// Local HTTP development wants "lax"; a production deployment where the
	// frontend and backend are on different domains needs "none", which
	// browsers only honour alongside CookieSecure=true.
	CookieSameSite string
	// FrontendOrigin is the exact origin allowed to make credentialed requests.
	// Cookie auth rules out a wildcard, so this cannot be "*".
	FrontendOrigin string
}

var (
	// ErrMissingJWTSecret is returned when JWT_ACCESS_SECRET is unset. Falling
	// back to a default would let every deployment sign tokens the others accept.
	ErrMissingJWTSecret = errors.New("JWT_ACCESS_SECRET must be set")
	// ErrInvalidCookieSameSite is returned when COOKIE_SAMESITE is set to
	// anything other than the three values browsers define.
	ErrInvalidCookieSameSite = errors.New(`COOKIE_SAMESITE must be "lax", "strict", or "none"`)
	// ErrCookieSameSiteNeedsSecure is returned for the one combination
	// browsers reject outright: SameSite=None cookies are dropped unless they
	// are also marked Secure.
	ErrCookieSameSiteNeedsSecure = errors.New("COOKIE_SAMESITE=none requires COOKIE_SECURE=true")
)

func Load() Config {
	_ = godotenv.Load()

	return Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://reload:reload@localhost:5432/reloaddb?sslmode=disable"),
		Port:            getEnv("PORT", "8080"),
		JWTAccessSecret: getEnv("JWT_ACCESS_SECRET", ""),
		CookieDomain:    getEnv("COOKIE_DOMAIN", ""),
		CookieSecure:    getEnv("COOKIE_SECURE", "false") == "true",
		CookieSameSite:  strings.ToLower(getEnv("COOKIE_SAMESITE", "lax")),
		FrontendOrigin:  getEnv("FRONTEND_ORIGIN", "http://localhost:3000"),
	}
}

// Validate reports configuration that would make the server unsafe or unable to
// serve auth, so startup can fail loudly instead of at the first login.
func (c Config) Validate() error {
	if c.JWTAccessSecret == "" {
		return ErrMissingJWTSecret
	}

	switch c.CookieSameSite {
	case "lax", "strict", "none":
	default:
		return fmt.Errorf("%w (got %q)", ErrInvalidCookieSameSite, c.CookieSameSite)
	}
	if c.CookieSameSite == "none" && !c.CookieSecure {
		return ErrCookieSameSiteNeedsSecure
	}

	return nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
