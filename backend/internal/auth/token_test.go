package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-not-used-anywhere-real"

func TestGenerateAndParseAccessToken(t *testing.T) {
	tm := NewTokenManager(testSecret)

	token, expiresAt, err := tm.GenerateAccessToken(42)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	if got := time.Until(expiresAt); got > AccessTokenTTL || got < AccessTokenTTL-time.Minute {
		t.Errorf("expiry %v is not roughly AccessTokenTTL (%v) away", got, AccessTokenTTL)
	}

	userID, err := tm.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("ParseAccessToken: %v", err)
	}
	if userID != 42 {
		t.Errorf("user ID = %d, want 42", userID)
	}
}

func TestParseAccessTokenRejectsOtherSecret(t *testing.T) {
	token, _, err := NewTokenManager(testSecret).GenerateAccessToken(1)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	// A token signed by one deployment must not be accepted by another.
	_, err = NewTokenManager("a-completely-different-secret").ParseAccessToken(token)
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("error = %v, want ErrInvalidToken", err)
	}
}

func TestParseAccessTokenRejectsMalformed(t *testing.T) {
	tm := NewTokenManager(testSecret)

	valid, _, err := tm.GenerateAccessToken(7)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	tests := map[string]string{
		"empty":            "",
		"not a jwt":        "hello",
		"missing segments": "abc.def",
		// Flipping the last character invalidates the signature.
		"tampered signature": valid[:len(valid)-1] + string(flipByte(valid[len(valid)-1])),
	}

	for name, token := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := tm.ParseAccessToken(token); !errors.Is(err, ErrInvalidToken) {
				t.Errorf("error = %v, want ErrInvalidToken", err)
			}
		})
	}
}

// TestParseAccessTokenRejectsUnsignedToken covers the "alg: none" confusion
// attack: a token whose header claims it needs no signature must not be
// trusted just because its claims parse.
func TestParseAccessTokenRejectsUnsignedToken(t *testing.T) {
	claims := AccessClaims{
		UserID: 99,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	unsigned, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).
		SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("signing unsigned token: %v", err)
	}

	if _, err := NewTokenManager(testSecret).ParseAccessToken(unsigned); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("error = %v, want ErrInvalidToken", err)
	}
}

func TestParseAccessTokenReportsExpirySeparately(t *testing.T) {
	// An expired token is well-formed and correctly signed, so it must map to
	// ErrExpiredToken — that is what tells the client a silent refresh is worth
	// attempting rather than a trip to the login page.
	claims := AccessClaims{
		UserID: 5,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}

	expired, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("signing expired token: %v", err)
	}

	_, err = NewTokenManager(testSecret).ParseAccessToken(expired)
	if !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("error = %v, want ErrExpiredToken", err)
	}
	if errors.Is(err, ErrInvalidToken) {
		t.Error("expired token must not also report as invalid")
	}
}

// TestParseAccessTokenRejectsZeroUserID guards the case where a correctly
// signed token carries no user_id: uint's zero value must not read as "user 0".
func TestParseAccessTokenRejectsZeroUserID(t *testing.T) {
	claims := jwt.RegisteredClaims{
		Subject:   "0",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("signing claimless token: %v", err)
	}

	if _, err := NewTokenManager(testSecret).ParseAccessToken(token); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("error = %v, want ErrInvalidToken", err)
	}
}

func TestGenerateRefreshTokenIsUnique(t *testing.T) {
	const iterations = 100
	seen := make(map[string]bool, iterations)

	for range iterations {
		token, expiresAt, err := GenerateRefreshToken()
		if err != nil {
			t.Fatalf("GenerateRefreshToken: %v", err)
		}
		if seen[token] {
			t.Fatalf("duplicate refresh token generated: %q", token)
		}
		seen[token] = true

		if token == "" {
			t.Fatal("refresh token is empty")
		}
		if got := time.Until(expiresAt); got > RefreshTokenTTL || got < RefreshTokenTTL-time.Minute {
			t.Errorf("expiry %v is not roughly RefreshTokenTTL (%v) away", got, RefreshTokenTTL)
		}
	}
}

func TestHashToken(t *testing.T) {
	token, _, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}

	hash := HashToken(token)

	if hash == token {
		t.Error("hash must not equal the raw token, or storing it protects nothing")
	}
	if HashToken(token) != hash {
		t.Error("hashing the same token twice must produce the same digest")
	}
	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64 hex characters", len(hash))
	}

	other, _, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	if HashToken(other) == hash {
		t.Error("distinct tokens hashed to the same digest")
	}
}

// flipByte returns a different character in the base64url alphabet, used to
// corrupt a signature without making the token structurally invalid.
func flipByte(b byte) byte {
	if b == 'A' {
		return 'B'
	}
	return 'A'
}
