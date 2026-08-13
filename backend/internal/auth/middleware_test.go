package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// runWithMiddleware sends a request through middleware to a handler that
// records whether it ran and what user ID it saw.
func runWithMiddleware(t *testing.T, middleware gin.HandlerFunc, cookie *http.Cookie) (*httptest.ResponseRecorder, bool, uint) {
	t.Helper()

	var (
		handlerRan bool
		seenUserID uint
	)

	router := gin.New()
	router.GET("/protected", middleware, func(c *gin.Context) {
		handlerRan = true
		seenUserID, _ = UserIDFromContext(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	return recorder, handlerRan, seenUserID
}

func accessCookie(t *testing.T, tm *TokenManager, userID uint) *http.Cookie {
	t.Helper()

	token, _, err := tm.GenerateAccessToken(userID)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	return &http.Cookie{Name: AccessCookieName, Value: token}
}

func expiredAccessCookie(t *testing.T, secret string) *http.Cookie {
	t.Helper()

	claims := AccessClaims{
		UserID: 3,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("signing expired token: %v", err)
	}
	return &http.Cookie{Name: AccessCookieName, Value: token}
}

// errorCode pulls the machine-readable code out of a 401 body.
func errorCode(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()

	var body struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decoding response body %q: %v", recorder.Body.String(), err)
	}
	return body.Code
}

func TestRequireAuthAllowsValidToken(t *testing.T) {
	tm := NewTokenManager(testSecret)

	recorder, handlerRan, userID := runWithMiddleware(t, RequireAuth(tm), accessCookie(t, tm, 17))

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", recorder.Code)
	}
	if !handlerRan {
		t.Fatal("handler did not run behind a valid token")
	}
	if userID != 17 {
		t.Errorf("user ID in context = %d, want 17", userID)
	}
}

func TestRequireAuthRejects(t *testing.T) {
	tm := NewTokenManager(testSecret)

	tests := []struct {
		name     string
		cookie   *http.Cookie
		wantCode string
	}{
		{
			name:     "no cookie",
			cookie:   nil,
			wantCode: CodeUnauthorized,
		},
		{
			name:     "empty cookie",
			cookie:   &http.Cookie{Name: AccessCookieName, Value: ""},
			wantCode: CodeUnauthorized,
		},
		{
			name:     "garbage token",
			cookie:   &http.Cookie{Name: AccessCookieName, Value: "not-a-jwt"},
			wantCode: CodeUnauthorized,
		},
		{
			// Signed by someone else — never a reason to try refreshing.
			name:     "foreign signature",
			cookie:   accessCookie(t, NewTokenManager("some-other-secret"), 17),
			wantCode: CodeUnauthorized,
		},
		{
			// The one case worth a silent refresh, so it gets its own code.
			name:     "expired token",
			cookie:   expiredAccessCookie(t, testSecret),
			wantCode: CodeTokenExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder, handlerRan, _ := runWithMiddleware(t, RequireAuth(tm), tt.cookie)

			if recorder.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", recorder.Code)
			}
			if handlerRan {
				t.Error("handler ran despite failed authentication")
			}
			if got := errorCode(t, recorder); got != tt.wantCode {
				t.Errorf("code = %q, want %q", got, tt.wantCode)
			}
		})
	}
}

// TestRequireAuthIgnoresRefreshCookie makes sure a refresh token cannot stand in
// for an access token: it is opaque, so it must not authenticate a request.
func TestRequireAuthIgnoresRefreshCookie(t *testing.T) {
	tm := NewTokenManager(testSecret)

	token, _, err := tm.GenerateAccessToken(17)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	recorder, handlerRan, _ := runWithMiddleware(t, RequireAuth(tm),
		&http.Cookie{Name: RefreshCookieName, Value: token})

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", recorder.Code)
	}
	if handlerRan {
		t.Error("handler ran on a request carrying only a refresh cookie")
	}
}

func TestOptionalAuthIdentifiesWhenItCan(t *testing.T) {
	tm := NewTokenManager(testSecret)

	recorder, handlerRan, userID := runWithMiddleware(t, OptionalAuth(tm), accessCookie(t, tm, 21))

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", recorder.Code)
	}
	if !handlerRan {
		t.Fatal("handler did not run")
	}
	if userID != 21 {
		t.Errorf("user ID in context = %d, want 21", userID)
	}
}

func TestOptionalAuthLetsAnonymousThrough(t *testing.T) {
	tm := NewTokenManager(testSecret)

	tests := map[string]*http.Cookie{
		"no cookie":     nil,
		"garbage token": {Name: AccessCookieName, Value: "not-a-jwt"},
		"expired token": expiredAccessCookie(t, testSecret),
	}

	for name, cookie := range tests {
		t.Run(name, func(t *testing.T) {
			recorder, handlerRan, userID := runWithMiddleware(t, OptionalAuth(tm), cookie)

			if recorder.Code != http.StatusOK {
				t.Errorf("status = %d, want 200", recorder.Code)
			}
			if !handlerRan {
				t.Fatal("handler did not run for an anonymous request")
			}
			// A bad token must leave the request anonymous rather than
			// half-authenticated.
			if userID != 0 {
				t.Errorf("user ID in context = %d, want 0", userID)
			}
		})
	}
}

func TestUserIDFromContext(t *testing.T) {
	tests := map[string]struct {
		set    bool
		value  any
		wantID uint
		wantOK bool
	}{
		"authenticated":  {set: true, value: uint(9), wantID: 9, wantOK: true},
		"never set":      {set: false},
		"zero user id":   {set: true, value: uint(0)},
		"wrong type":     {set: true, value: "9"},
		"signed integer": {set: true, value: 9},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c, _ := newTestContext()
			if tt.set {
				c.Set(ContextUserIDKey, tt.value)
			}

			userID, ok := UserIDFromContext(c)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if userID != tt.wantID {
				t.Errorf("user ID = %d, want %d", userID, tt.wantID)
			}
		})
	}
}
