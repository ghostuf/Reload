package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestContext returns a gin context writing to a fresh recorder, plus the
// recorder, for exercising the cookie helpers in isolation.
func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c, recorder
}

// cookiesByName parses the Set-Cookie headers on a recorded response.
func cookiesByName(t *testing.T, recorder *httptest.ResponseRecorder) map[string]*http.Cookie {
	t.Helper()

	byName := make(map[string]*http.Cookie)
	for _, cookie := range recorder.Result().Cookies() {
		byName[cookie.Name] = cookie
	}
	return byName
}

func TestSetAuthCookies(t *testing.T) {
	c, recorder := newTestContext()

	SetAuthCookies(c, CookieOptions{Domain: "example.com", Secure: true, SameSite: http.SameSiteLaxMode}, "access-value", "refresh-value")

	cookies := cookiesByName(t, recorder)

	access, ok := cookies[AccessCookieName]
	if !ok {
		t.Fatalf("no %s cookie was set", AccessCookieName)
	}
	refresh, ok := cookies[RefreshCookieName]
	if !ok {
		t.Fatalf("no %s cookie was set", RefreshCookieName)
	}

	if access.Value != "access-value" {
		t.Errorf("access cookie value = %q, want %q", access.Value, "access-value")
	}
	if refresh.Value != "refresh-value" {
		t.Errorf("refresh cookie value = %q, want %q", refresh.Value, "refresh-value")
	}

	// Both tokens are bearer credentials: script must never be able to read them.
	for _, cookie := range []*http.Cookie{access, refresh} {
		if !cookie.HttpOnly {
			t.Errorf("%s is not HttpOnly", cookie.Name)
		}
		if !cookie.Secure {
			t.Errorf("%s is not Secure despite Secure: true", cookie.Name)
		}
		if cookie.Domain != "example.com" {
			t.Errorf("%s domain = %q, want example.com", cookie.Name, cookie.Domain)
		}
		if cookie.SameSite != http.SameSiteLaxMode {
			t.Errorf("%s SameSite = %v, want Lax", cookie.Name, cookie.SameSite)
		}
	}

	// The whole point of the split: the refresh token is scoped to /auth so it
	// is not attached to ordinary API calls.
	if access.Path != "/" {
		t.Errorf("access cookie path = %q, want /", access.Path)
	}
	if refresh.Path != "/auth" {
		t.Errorf("refresh cookie path = %q, want /auth", refresh.Path)
	}

	if access.MaxAge != int(AccessTokenTTL.Seconds()) {
		t.Errorf("access cookie MaxAge = %d, want %d", access.MaxAge, int(AccessTokenTTL.Seconds()))
	}
	if refresh.MaxAge != int(RefreshTokenTTL.Seconds()) {
		t.Errorf("refresh cookie MaxAge = %d, want %d", refresh.MaxAge, int(RefreshTokenTTL.Seconds()))
	}
}

func TestSetAuthCookiesInsecureForLocalDev(t *testing.T) {
	c, recorder := newTestContext()

	SetAuthCookies(c, CookieOptions{}, "access-value", "refresh-value")

	for name, cookie := range cookiesByName(t, recorder) {
		if cookie.Secure {
			t.Errorf("%s is Secure despite Secure: false", name)
		}
		if cookie.Domain != "" {
			t.Errorf("%s domain = %q, want empty (host-scoped)", name, cookie.Domain)
		}
		if !cookie.HttpOnly {
			t.Errorf("%s is not HttpOnly", name)
		}
	}
}

func TestSetAccessCookieLeavesRefreshAlone(t *testing.T) {
	c, recorder := newTestContext()

	SetAccessCookie(c, CookieOptions{}, "access-only")

	cookies := cookiesByName(t, recorder)
	if _, ok := cookies[AccessCookieName]; !ok {
		t.Errorf("no %s cookie was set", AccessCookieName)
	}
	if _, ok := cookies[RefreshCookieName]; ok {
		t.Errorf("%s was rewritten by SetAccessCookie", RefreshCookieName)
	}
}

func TestClearAuthCookies(t *testing.T) {
	c, recorder := newTestContext()

	ClearAuthCookies(c, CookieOptions{Domain: "example.com", Secure: true})

	cookies := cookiesByName(t, recorder)

	for _, name := range []string{AccessCookieName, RefreshCookieName} {
		cookie, ok := cookies[name]
		if !ok {
			t.Fatalf("%s was not cleared", name)
		}
		if cookie.Value != "" {
			t.Errorf("%s value = %q, want empty", name, cookie.Value)
		}
		if cookie.MaxAge >= 0 {
			t.Errorf("%s MaxAge = %d, want negative so the browser drops it", name, cookie.MaxAge)
		}
	}

	// A browser only replaces a cookie when name, path, and domain all match,
	// so clearing has to reuse the paths the cookies were written with.
	if got := cookies[AccessCookieName].Path; got != "/" {
		t.Errorf("cleared access cookie path = %q, want /", got)
	}
	if got := cookies[RefreshCookieName].Path; got != "/auth" {
		t.Errorf("cleared refresh cookie path = %q, want /auth", got)
	}
	if got := cookies[RefreshCookieName].Domain; got != "example.com" {
		t.Errorf("cleared refresh cookie domain = %q, want example.com", got)
	}
}

func TestReadCookies(t *testing.T) {
	tests := []struct {
		name      string
		cookie    *http.Cookie
		read      func(*gin.Context) (string, bool)
		wantValue string
		wantOK    bool
	}{
		{
			name:      "access present",
			cookie:    &http.Cookie{Name: AccessCookieName, Value: "access-value"},
			read:      ReadAccessCookie,
			wantValue: "access-value",
			wantOK:    true,
		},
		{
			name:      "refresh present",
			cookie:    &http.Cookie{Name: RefreshCookieName, Value: "refresh-value"},
			read:      ReadRefreshCookie,
			wantValue: "refresh-value",
			wantOK:    true,
		},
		{
			name:   "access missing",
			cookie: nil,
			read:   ReadAccessCookie,
		},
		{
			name:   "refresh missing",
			cookie: nil,
			read:   ReadRefreshCookie,
		},
		{
			// An empty cookie is as good as no cookie; it must not read as a
			// token the parser then has to reject.
			name:   "access empty",
			cookie: &http.Cookie{Name: AccessCookieName, Value: ""},
			read:   ReadAccessCookie,
		},
		{
			name:   "refresh empty",
			cookie: &http.Cookie{Name: RefreshCookieName, Value: ""},
			read:   ReadRefreshCookie,
		},
		{
			// The refresh reader must not fall back to the access cookie.
			name:   "wrong cookie name",
			cookie: &http.Cookie{Name: AccessCookieName, Value: "access-value"},
			read:   ReadRefreshCookie,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := newTestContext()
			if tt.cookie != nil {
				c.Request.AddCookie(tt.cookie)
			}

			value, ok := tt.read(c)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if value != tt.wantValue {
				t.Errorf("value = %q, want %q", value, tt.wantValue)
			}
		})
	}
}
