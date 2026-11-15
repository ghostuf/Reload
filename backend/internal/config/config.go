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
	// FrontendURL is the frontend's public URL, used to build the
	// success_url/failure_url eSewa redirects back to after payment. Usually
	// the same value as FrontendOrigin; kept separate because one is a CORS
	// match and the other is a URL to redirect to, and they only coincide by
	// convention rather than necessity.
	FrontendURL string

	// eSewa ePay v2 (sandbox/test mode). ESewaSecretKey must never reach the
	// frontend -- it is used only here, server-side, to sign the outgoing
	// request and verify eSewa's response.
	ESewaEnv         string
	ESewaProductCode string
	ESewaSecretKey   string
	ESewaFormURL     string
	ESewaStatusURL   string

	// PaymentMode is "mock" or "esewa" and decides which checkout flow
	// POST /checkout and POST /checkout/esewa/initiate will actually serve
	// -- GET /config/payment reports it so the frontend knows which one to
	// use. Kept switchable so a sandbox outage never has to block a demo.
	PaymentMode string
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
	// ErrInvalidPaymentMode is returned when PAYMENT_MODE is set to anything
	// other than the two checkout flows this app actually has.
	ErrInvalidPaymentMode = errors.New(`PAYMENT_MODE must be "mock" or "esewa"`)
)

func Load() Config {
	_ = godotenv.Load()

	esewaEnv := strings.ToLower(getEnv("ESEWA_ENV", "sandbox"))

	return Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://reload:reload@localhost:5432/reloaddb?sslmode=disable"),
		Port:            getEnv("PORT", "8080"),
		JWTAccessSecret: getEnv("JWT_ACCESS_SECRET", ""),
		CookieDomain:    getEnv("COOKIE_DOMAIN", ""),
		CookieSecure:    getEnv("COOKIE_SECURE", "false") == "true",
		CookieSameSite:  strings.ToLower(getEnv("COOKIE_SAMESITE", "lax")),
		FrontendOrigin:  getEnv("FRONTEND_ORIGIN", "http://localhost:3000"),
		FrontendURL:     getEnv("FRONTEND_URL", "http://localhost:3000"),

		ESewaEnv: esewaEnv,
		// EPAYTEST and the UAT key below are eSewa's own published sandbox
		// values, not secrets belonging to this app, so they default in for
		// a working out-of-the-box demo. A production deployment must
		// override all three with the real merchant values eSewa issues.
		ESewaProductCode: getEnv("ESEWA_PRODUCT_CODE", "EPAYTEST"),
		ESewaSecretKey:   getEnv("ESEWA_SECRET_KEY", "8gBm/:&EnhH.1/q"),
		ESewaFormURL:     getEnv("ESEWA_FORM_URL", defaultEsewaFormURL(esewaEnv)),
		ESewaStatusURL:   getEnv("ESEWA_STATUS_URL", defaultEsewaStatusURL(esewaEnv)),

		PaymentMode: strings.ToLower(getEnv("PAYMENT_MODE", defaultPaymentMode(getEnv("ESEWA_SECRET_KEY", "8gBm/:&EnhH.1/q")))),
	}
}

// defaultPaymentMode makes PAYMENT_MODE default to "esewa" exactly when the
// task describes: when the sandbox (or a real merchant's) secret key is
// actually set. An explicit PAYMENT_MODE always overrides this.
func defaultPaymentMode(esewaSecretKey string) string {
	if esewaSecretKey == "" {
		return "mock"
	}
	return "esewa"
}

// defaultEsewaFormURL and defaultEsewaStatusURL pick eSewa's sandbox
// ("rc-"-prefixed) or production endpoint based on ESEWA_ENV. They are only
// consulted when ESEWA_FORM_URL / ESEWA_STATUS_URL are not set explicitly,
// which always takes precedence.
func defaultEsewaFormURL(env string) string {
	if env == "production" {
		return "https://epay.esewa.com.np/api/epay/main/v2/form"
	}
	return "https://rc-epay.esewa.com.np/api/epay/main/v2/form"
}

func defaultEsewaStatusURL(env string) string {
	if env == "production" {
		return "https://esewa.com.np/api/epay/transaction/status/"
	}
	return "https://rc.esewa.com.np/api/epay/transaction/status/"
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

	switch c.PaymentMode {
	case "mock", "esewa":
	default:
		return fmt.Errorf("%w (got %q)", ErrInvalidPaymentMode, c.PaymentMode)
	}

	return nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
