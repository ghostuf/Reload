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

// CookieOptions controls how auth cookies are written. Secure should be true in
// any environment served over HTTPS.
type CookieOptions struct {
	Domain string
	Secure bool
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
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(name, value, maxAge, path, opts.Domain, opts.Secure, true)
}
