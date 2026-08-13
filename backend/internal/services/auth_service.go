package services

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/ghostuf/reload/internal/auth"
	"github.com/ghostuf/reload/internal/models"
	"github.com/ghostuf/reload/internal/repository"
	"gorm.io/gorm"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 72 // bcrypt silently ignores bytes past 72
	maxDisplayName    = 60
)

var (
	ErrInvalidEmail        = errors.New("email is not valid")
	ErrWeakPassword        = fmt.Errorf("password must be between %d and %d characters", minPasswordLength, maxPasswordLength)
	ErrDisplayNameTooLong  = fmt.Errorf("display name must be at most %d characters", maxDisplayName)
	ErrEmailTaken          = errors.New("email is already registered")
	ErrInvalidCredentials  = errors.New("email or password is incorrect")
	ErrInvalidRefreshToken = errors.New("refresh token is missing, expired, or already used")
)

// dummyHash is a valid bcrypt hash of a value nobody knows. Verifying against
// it on a missing-user login keeps the response time comparable to a real
// failed login, so the endpoint does not leak which emails are registered.
const dummyHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

type RegisterParams struct {
	Email       string
	Password    string
	DisplayName string
}

// TokenPair carries the raw tokens for a session. The refresh value exists in
// this form only long enough to be written to a cookie.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// UserStore and RefreshTokenStore are the slices of persistence this service
// needs. They are interfaces so the auth rules can be exercised against
// in-memory fakes; *repository.UserRepository and *repository.RefreshTokenRepository
// are the production implementations.
type UserStore interface {
	Create(user *models.User) error
	GetByEmail(email string) (*models.User, error)
	GetByID(id uint) (*models.User, error)
}

type RefreshTokenStore interface {
	Create(token *models.RefreshToken) error
	GetByHash(hash string) (*models.RefreshToken, error)
	Revoke(id uint) error
	RevokeAllForUser(userID uint) error
	Rotate(oldID uint, replacement *models.RefreshToken) error
}

// Fail at compile time if the repositories drift away from what this needs.
var (
	_ UserStore         = (*repository.UserRepository)(nil)
	_ RefreshTokenStore = (*repository.RefreshTokenRepository)(nil)
)

type AuthService struct {
	users  UserStore
	tokens RefreshTokenStore
	tm     *auth.TokenManager
}

func NewAuthService(users UserStore, tokens RefreshTokenStore, tm *auth.TokenManager) *AuthService {
	return &AuthService{users: users, tokens: tokens, tm: tm}
}

// Register validates the input, hashes the password, and creates the user.
func (s *AuthService) Register(params RegisterParams) (*models.User, error) {
	email, err := normaliseEmail(params.Email)
	if err != nil {
		return nil, err
	}
	if err := validatePassword(params.Password); err != nil {
		return nil, err
	}

	displayName := strings.TrimSpace(params.DisplayName)
	if len([]rune(displayName)) > maxDisplayName {
		return nil, ErrDisplayNameTooLong
	}
	if displayName == "" {
		displayName = email[:strings.Index(email, "@")]
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        email,
		PasswordHash: hash,
		DisplayName:  displayName,
	}

	// The unique index is the source of truth: checking first and inserting
	// after would still race with a concurrent signup.
	if err := s.users.Create(user); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("creating user: %w", err)
	}

	return user, nil
}

// Login verifies credentials and issues a fresh token pair.
func (s *AuthService) Login(email, password string) (*models.User, TokenPair, error) {
	normalised, err := normaliseEmail(email)
	if err != nil {
		return nil, TokenPair{}, ErrInvalidCredentials
	}

	user, err := s.users.GetByEmail(normalised)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Spend the same work as a real verification before failing.
			_ = auth.VerifyPassword(dummyHash, password)
			return nil, TokenPair{}, ErrInvalidCredentials
		}
		return nil, TokenPair{}, fmt.Errorf("looking up user: %w", err)
	}

	if err := auth.VerifyPassword(user.PasswordHash, password); err != nil {
		if errors.Is(err, auth.ErrPasswordMismatch) {
			return nil, TokenPair{}, ErrInvalidCredentials
		}
		return nil, TokenPair{}, fmt.Errorf("verifying password: %w", err)
	}

	pair, err := s.issueTokens(user.ID)
	if err != nil {
		return nil, TokenPair{}, err
	}

	return user, pair, nil
}

// Refresh validates a raw refresh token and rotates it, returning a new pair.
// The presented token is revoked whether or not rotation succeeds.
func (s *AuthService) Refresh(rawToken string) (TokenPair, error) {
	if rawToken == "" {
		return TokenPair{}, ErrInvalidRefreshToken
	}

	stored, err := s.tokens.GetByHash(auth.HashToken(rawToken))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return TokenPair{}, ErrInvalidRefreshToken
		}
		return TokenPair{}, fmt.Errorf("looking up refresh token: %w", err)
	}

	// A revoked token being presented again means either a replay of a stolen
	// token or a leaked one. Either way the session family is no longer
	// trustworthy, so tear all of it down.
	if stored.Revoked {
		if err := s.tokens.RevokeAllForUser(stored.UserID); err != nil {
			return TokenPair{}, err
		}
		return TokenPair{}, ErrInvalidRefreshToken
	}

	if time.Now().After(stored.ExpiresAt) {
		return TokenPair{}, ErrInvalidRefreshToken
	}

	rawRefresh, expiresAt, err := auth.GenerateRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}

	replacement := &models.RefreshToken{
		UserID:    stored.UserID,
		TokenHash: auth.HashToken(rawRefresh),
		ExpiresAt: expiresAt,
	}

	if err := s.tokens.Rotate(stored.ID, replacement); err != nil {
		// Rotate reports not-found when a concurrent request already consumed
		// this token.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return TokenPair{}, ErrInvalidRefreshToken
		}
		return TokenPair{}, err
	}

	accessToken, _, err := s.tm.GenerateAccessToken(stored.UserID)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{AccessToken: accessToken, RefreshToken: rawRefresh}, nil
}

// Logout revokes the presented refresh token. It is deliberately forgiving: an
// unknown or already-revoked token still counts as a successful logout, since
// the caller's cookies get cleared either way.
func (s *AuthService) Logout(rawToken string) error {
	if rawToken == "" {
		return nil
	}

	stored, err := s.tokens.GetByHash(auth.HashToken(rawToken))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("looking up refresh token: %w", err)
	}

	return s.tokens.Revoke(stored.ID)
}

// GetUser returns the user behind an authenticated request.
func (s *AuthService) GetUser(id uint) (*models.User, error) {
	return s.users.GetByID(id)
}

// issueTokens mints an access token and a stored refresh token for a session.
func (s *AuthService) issueTokens(userID uint) (TokenPair, error) {
	accessToken, _, err := s.tm.GenerateAccessToken(userID)
	if err != nil {
		return TokenPair{}, err
	}

	rawRefresh, expiresAt, err := auth.GenerateRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}

	record := &models.RefreshToken{
		UserID:    userID,
		TokenHash: auth.HashToken(rawRefresh),
		ExpiresAt: expiresAt,
	}
	if err := s.tokens.Create(record); err != nil {
		return TokenPair{}, fmt.Errorf("storing refresh token: %w", err)
	}

	return TokenPair{AccessToken: accessToken, RefreshToken: rawRefresh}, nil
}

func normaliseEmail(email string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(email))
	if trimmed == "" {
		return "", ErrInvalidEmail
	}

	addr, err := mail.ParseAddress(trimmed)
	if err != nil {
		return "", ErrInvalidEmail
	}
	// ParseAddress accepts "Name <a@b.com>"; only the bare address is wanted.
	if addr.Address != trimmed || !strings.Contains(trimmed, "@") {
		return "", ErrInvalidEmail
	}

	return trimmed, nil
}

func validatePassword(password string) error {
	if len(password) < minPasswordLength || len(password) > maxPasswordLength {
		return ErrWeakPassword
	}
	return nil
}
