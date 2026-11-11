package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	// AccessCookieName holds the signed JWT access token.
	AccessCookieName = "reload_access_token"
	// RefreshCookieName holds the opaque refresh token.
	RefreshCookieName = "reload_refresh_token"

	// accessCookiePath is "/" because every protected route needs the access
	// token.
	accessCookiePath = "/"
	// refreshCookiePath scopes the refresh token to the endpoints that consume
	// it, so it is not attached to ordinary API requests.
	refreshCookiePath = "/auth"
)

// CookieOptions controls how auth cookies are written. Secure should be true
// in any environment served over HTTPS, and SameSite should be
// http.SameSiteNoneMode (alongside Secure) whenever the frontend and backend
// are on different domains -- otherwise the browser never attaches the
// cookie to the cross-site requests the frontend makes.
type CookieOptions struct {
	Domain   string
	Secure   bool
	SameSite http.SameSite
}

// ParseSameSite maps a config string onto http's SameSite enum. It defaults
// to Lax for anything unrecognised; config.Config.Validate is what actually
// rejects a bad value at startup, so by the time this runs s is guaranteed to
// be "lax", "strict", or "none".
func ParseSameSite(s string) http.SameSite {
	switch s {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

// SetAuthCookies writes both auth tokens as httpOnly cookies.
func SetAuthCookies(c *gin.Context, opts CookieOptions, accessToken, refreshToken string) {
	SetAccessCookie(c, opts, accessToken)
	setCookie(c, RefreshCookieName, refreshToken, int(RefreshTokenTTL.Seconds()), refreshCookiePath, opts)
}

// SetAccessCookie writes only the access token cookie, for refreshes that do
// not rotate the refresh token.
func SetAccessCookie(c *gin.Context, opts CookieOptions, accessToken string) {
	setCookie(c, AccessCookieName, accessToken, int(AccessTokenTTL.Seconds()), accessCookiePath, opts)
}

// ClearAuthCookies expires both auth cookies. The path and domain must match
// the ones used when setting them, or the browser will keep the originals.
func ClearAuthCookies(c *gin.Context, opts CookieOptions) {
	setCookie(c, AccessCookieName, "", -1, accessCookiePath, opts)
	setCookie(c, RefreshCookieName, "", -1, refreshCookiePath, opts)
}

// ReadRefreshCookie returns the raw refresh token from the request, if present.
func ReadRefreshCookie(c *gin.Context) (string, bool) {
	token, err := c.Cookie(RefreshCookieName)
	if err != nil || token == "" {
		return "", false
	}
	return token, true
}

// ReadAccessCookie returns the raw access token from the request, if present.
func ReadAccessCookie(c *gin.Context) (string, bool) {
	token, err := c.Cookie(AccessCookieName)
	if err != nil || token == "" {
		return "", false
	}
	return token, true
}

func setCookie(c *gin.Context, name, value string, maxAge int, path string, opts CookieOptions) {
	c.SetSameSite(opts.SameSite)
	c.SetCookie(name, value, maxAge, path, opts.Domain, opts.Secure, true)
}
