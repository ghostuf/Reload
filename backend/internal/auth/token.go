package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// AccessTokenTTL is how long a signed access token stays valid.
	AccessTokenTTL = 15 * time.Minute
	// RefreshTokenTTL is how long an opaque refresh token stays valid.
	RefreshTokenTTL = 7 * 24 * time.Hour

	// refreshTokenBytes is the entropy of a raw refresh token before encoding.
	refreshTokenBytes = 32
)

var (
	// ErrInvalidToken is returned when an access token is malformed, has an
	// unexpected signing method, or fails signature verification.
	ErrInvalidToken = errors.New("invalid access token")
	// ErrExpiredToken is returned when an access token is well-formed and
	// correctly signed but past its expiry.
	ErrExpiredToken = errors.New("expired access token")
)

// AccessClaims are the claims carried by an access token.
type AccessClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// TokenManager issues and verifies access tokens, and mints refresh tokens.
type TokenManager struct {
	accessSecret []byte
}

// NewTokenManager returns a TokenManager signing access tokens with secret.
func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{accessSecret: []byte(secret)}
}

// GenerateAccessToken returns a signed JWT carrying userID, along with its
// expiry time.
func (m *TokenManager) GenerateAccessToken(userID uint) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(AccessTokenTTL)

	claims := AccessClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprint(userID),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.accessSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("signing access token: %w", err)
	}

	return signed, expiresAt, nil
}

// ParseAccessToken verifies the signature and expiry of an access token and
// returns the user ID it carries. It returns ErrExpiredToken for a valid but
// expired token, and ErrInvalidToken for everything else.
func (m *TokenManager) ParseAccessToken(token string) (uint, error) {
	var claims AccessClaims

	_, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %q", t.Header["alg"])
		}
		return m.accessSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, ErrExpiredToken
		}
		return 0, ErrInvalidToken
	}

	if claims.UserID == 0 {
		return 0, ErrInvalidToken
	}

	return claims.UserID, nil
}

// GenerateRefreshToken returns a cryptographically random opaque token and its
// expiry. The raw value is shown to the client exactly once; only HashToken of
// it is ever persisted.
func GenerateRefreshToken() (string, time.Time, error) {
	buf := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", time.Time{}, fmt.Errorf("generating refresh token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(buf), time.Now().Add(RefreshTokenTTL), nil
}

// HashToken returns the hex-encoded SHA-256 of an opaque token. Refresh tokens
// are full-entropy random values, so a fast digest is appropriate here — unlike
// passwords, they are not guessable and need no key stretching.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
