package services

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ghostuf/reload/internal/auth"
	"github.com/ghostuf/reload/internal/models"
	"gorm.io/gorm"
)

const testSecret = "test-secret-not-used-anywhere-real"

// fakeUserStore is an in-memory UserStore that reproduces the behaviour the
// service depends on: a unique index on email, and gorm's sentinel errors.
type fakeUserStore struct {
	users  map[uint]*models.User
	nextID uint
	// createErr and lookupErr, when set, stand in for an unreachable database.
	createErr error
	lookupErr error
}

func newFakeUserStore() *fakeUserStore {
	return &fakeUserStore{users: make(map[uint]*models.User), nextID: 1}
}

func (s *fakeUserStore) Create(user *models.User) error {
	if s.createErr != nil {
		return s.createErr
	}
	for _, existing := range s.users {
		if existing.Email == user.Email {
			return gorm.ErrDuplicatedKey
		}
	}

	user.ID = s.nextID
	s.nextID++
	user.CreatedAt = time.Now()
	// Store a copy so later mutations by the caller cannot rewrite history.
	stored := *user
	s.users[user.ID] = &stored
	return nil
}

func (s *fakeUserStore) GetByEmail(email string) (*models.User, error) {
	if s.lookupErr != nil {
		return nil, s.lookupErr
	}
	for _, user := range s.users {
		if user.Email == email {
			copied := *user
			return &copied, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *fakeUserStore) GetByID(id uint) (*models.User, error) {
	user, ok := s.users[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	copied := *user
	return &copied, nil
}

// fakeTokenStore is an in-memory RefreshTokenStore. Rotate mirrors the
// repository's transactional guarantee: revoke-then-insert, refusing to act on
// an already-revoked row.
type fakeTokenStore struct {
	tokens map[uint]*models.RefreshToken
	nextID uint
	// rotations counts successful rotations, to assert a replay did not rotate.
	rotations int
	// createErr, lookupErr, and rotateErr stand in for database failures.
	createErr error
	lookupErr error
	rotateErr error
}

func newFakeTokenStore() *fakeTokenStore {
	return &fakeTokenStore{tokens: make(map[uint]*models.RefreshToken), nextID: 1}
}

func (s *fakeTokenStore) Create(token *models.RefreshToken) error {
	if s.createErr != nil {
		return s.createErr
	}
	token.ID = s.nextID
	s.nextID++
	stored := *token
	s.tokens[token.ID] = &stored
	return nil
}

func (s *fakeTokenStore) GetByHash(hash string) (*models.RefreshToken, error) {
	if s.lookupErr != nil {
		return nil, s.lookupErr
	}
	for _, token := range s.tokens {
		if token.TokenHash == hash {
			copied := *token
			return &copied, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *fakeTokenStore) Revoke(id uint) error {
	if token, ok := s.tokens[id]; ok {
		token.Revoked = true
	}
	return nil
}

func (s *fakeTokenStore) RevokeAllForUser(userID uint) error {
	for _, token := range s.tokens {
		if token.UserID == userID {
			token.Revoked = true
		}
	}
	return nil
}

func (s *fakeTokenStore) Rotate(oldID uint, replacement *models.RefreshToken) error {
	if s.rotateErr != nil {
		return s.rotateErr
	}
	old, ok := s.tokens[oldID]
	if !ok || old.Revoked {
		return gorm.ErrRecordNotFound
	}

	old.Revoked = true
	s.rotations++
	return s.Create(replacement)
}

// liveTokensFor counts a user's unrevoked, unexpired tokens.
func (s *fakeTokenStore) liveTokensFor(userID uint) int {
	live := 0
	for _, token := range s.tokens {
		if token.UserID == userID && !token.Revoked && time.Now().Before(token.ExpiresAt) {
			live++
		}
	}
	return live
}

func newTestService(t *testing.T) (*AuthService, *fakeUserStore, *fakeTokenStore) {
	t.Helper()

	users := newFakeUserStore()
	tokens := newFakeTokenStore()
	return NewAuthService(users, tokens, auth.NewTokenManager(testSecret)), users, tokens
}

// registerAndLogin creates a user and returns it alongside a live token pair.
func registerAndLogin(t *testing.T, s *AuthService, email, password string) (*models.User, TokenPair) {
	t.Helper()

	if _, err := s.Register(RegisterParams{Email: email, Password: password}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	user, pair, err := s.Login(email, password)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	return user, pair
}

func TestRegister(t *testing.T) {
	s, users, _ := newTestService(t)

	user, err := s.Register(RegisterParams{
		Email:       "  Player@Example.COM  ",
		Password:    "a good enough password",
		DisplayName: "  Player One  ",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Email is normalised so "A@b.com" and "a@b.com" cannot become two accounts.
	if user.Email != "player@example.com" {
		t.Errorf("email = %q, want player@example.com", user.Email)
	}
	if user.DisplayName != "Player One" {
		t.Errorf("display name = %q, want %q", user.DisplayName, "Player One")
	}
	if user.ID == 0 {
		t.Error("user was not assigned an ID")
	}

	stored, err := users.GetByEmail("player@example.com")
	if err != nil {
		t.Fatalf("stored user not found: %v", err)
	}
	if stored.PasswordHash == "a good enough password" {
		t.Fatal("password was stored in plaintext")
	}
	if err := auth.VerifyPassword(stored.PasswordHash, "a good enough password"); err != nil {
		t.Errorf("stored hash does not verify the password: %v", err)
	}
}

func TestRegisterDefaultsDisplayNameToEmailLocalPart(t *testing.T) {
	s, _, _ := newTestService(t)

	user, err := s.Register(RegisterParams{
		Email:    "solo.player@example.com",
		Password: "a good enough password",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	if user.DisplayName != "solo.player" {
		t.Errorf("display name = %q, want solo.player", user.DisplayName)
	}
}

func TestRegisterRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		params  RegisterParams
		wantErr error
	}{
		{
			name:    "empty email",
			params:  RegisterParams{Email: "  ", Password: "a good enough password"},
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "no at sign",
			params:  RegisterParams{Email: "not-an-email", Password: "a good enough password"},
			wantErr: ErrInvalidEmail,
		},
		{
			// mail.ParseAddress accepts this form; only a bare address is wanted.
			name:    "display form address",
			params:  RegisterParams{Email: "Player <player@example.com>", Password: "a good enough password"},
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "short password",
			params:  RegisterParams{Email: "player@example.com", Password: "short"},
			wantErr: ErrWeakPassword,
		},
		{
			name:    "empty password",
			params:  RegisterParams{Email: "player@example.com", Password: ""},
			wantErr: ErrWeakPassword,
		},
		{
			// Past 72 bytes bcrypt ignores the tail, so the service refuses it
			// rather than silently authenticating on a prefix.
			name:    "password past bcrypt limit",
			params:  RegisterParams{Email: "player@example.com", Password: strings.Repeat("a", 73)},
			wantErr: ErrWeakPassword,
		},
		{
			name: "display name too long",
			params: RegisterParams{
				Email:       "player@example.com",
				Password:    "a good enough password",
				DisplayName: strings.Repeat("x", 61),
			},
			wantErr: ErrDisplayNameTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, users, _ := newTestService(t)

			if _, err := s.Register(tt.params); !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
			if len(users.users) != 0 {
				t.Error("a user was created despite invalid input")
			}
		})
	}
}

func TestRegisterRejectsDuplicateEmail(t *testing.T) {
	s, _, _ := newTestService(t)

	if _, err := s.Register(RegisterParams{Email: "player@example.com", Password: "a good enough password"}); err != nil {
		t.Fatalf("first Register: %v", err)
	}

	// Different casing and padding must still collide with the stored address.
	_, err := s.Register(RegisterParams{Email: " PLAYER@example.com ", Password: "another fine password"})
	if !errors.Is(err, ErrEmailTaken) {
		t.Errorf("error = %v, want ErrEmailTaken", err)
	}
}

func TestLogin(t *testing.T) {
	s, _, tokens := newTestService(t)

	user, pair := registerAndLogin(t, s, "player@example.com", "a good enough password")

	if user.ID == 0 {
		t.Error("login returned a user without an ID")
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("login did not issue a full token pair")
	}

	userID, err := auth.NewTokenManager(testSecret).ParseAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("issued access token does not parse: %v", err)
	}
	if userID != user.ID {
		t.Errorf("access token carries user %d, want %d", userID, user.ID)
	}

	// Only the digest is persisted, so a database leak yields nothing usable.
	stored, err := tokens.GetByHash(auth.HashToken(pair.RefreshToken))
	if err != nil {
		t.Fatalf("refresh token was not stored: %v", err)
	}
	if stored.TokenHash == pair.RefreshToken {
		t.Error("refresh token was stored in its raw form")
	}
	if stored.Revoked {
		t.Error("freshly issued refresh token is already revoked")
	}
}

func TestLoginNormalisesEmail(t *testing.T) {
	s, _, _ := newTestService(t)

	if _, err := s.Register(RegisterParams{Email: "player@example.com", Password: "a good enough password"}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if _, _, err := s.Login("  PLAYER@Example.com  ", "a good enough password"); err != nil {
		t.Errorf("login with differently-cased email: %v", err)
	}
}

func TestLoginRejectsBadCredentials(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
	}{
		{name: "wrong password", email: "player@example.com", password: "not the password"},
		{name: "unknown email", email: "nobody@example.com", password: "a good enough password"},
		{name: "empty password", email: "player@example.com", password: ""},
		{name: "malformed email", email: "not-an-email", password: "a good enough password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _, tokens := newTestService(t)
			if _, err := s.Register(RegisterParams{Email: "player@example.com", Password: "a good enough password"}); err != nil {
				t.Fatalf("Register: %v", err)
			}

			_, pair, err := s.Login(tt.email, tt.password)
			// One error for every failure mode: a distinct "no such user"
			// would turn the endpoint into an account enumerator.
			if !errors.Is(err, ErrInvalidCredentials) {
				t.Errorf("error = %v, want ErrInvalidCredentials", err)
			}
			if pair.AccessToken != "" || pair.RefreshToken != "" {
				t.Error("tokens were issued for a failed login")
			}
			if len(tokens.tokens) != 0 {
				t.Error("a refresh token was stored for a failed login")
			}
		})
	}
}

func TestRefreshRotates(t *testing.T) {
	s, _, tokens := newTestService(t)

	user, pair := registerAndLogin(t, s, "player@example.com", "a good enough password")

	rotated, err := s.Refresh(pair.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}

	if rotated.RefreshToken == pair.RefreshToken {
		t.Error("refresh returned the same token instead of rotating it")
	}
	if rotated.AccessToken == "" {
		t.Error("refresh did not issue a new access token")
	}

	old, err := tokens.GetByHash(auth.HashToken(pair.RefreshToken))
	if err != nil {
		t.Fatalf("old token vanished instead of being revoked: %v", err)
	}
	if !old.Revoked {
		t.Error("the presented token was left live after rotation")
	}

	// Exactly one usable token per session, or a stolen one keeps working
	// alongside the legitimate one.
	if live := tokens.liveTokensFor(user.ID); live != 1 {
		t.Errorf("live tokens after rotation = %d, want 1", live)
	}

	userID, err := auth.NewTokenManager(testSecret).ParseAccessToken(rotated.AccessToken)
	if err != nil {
		t.Fatalf("rotated access token does not parse: %v", err)
	}
	if userID != user.ID {
		t.Errorf("rotated access token carries user %d, want %d", userID, user.ID)
	}
}

func TestRefreshChains(t *testing.T) {
	s, _, tokens := newTestService(t)

	user, pair := registerAndLogin(t, s, "player@example.com", "a good enough password")

	current := pair.RefreshToken
	for i := range 5 {
		next, err := s.Refresh(current)
		if err != nil {
			t.Fatalf("Refresh %d: %v", i+1, err)
		}
		if next.RefreshToken == current {
			t.Fatalf("refresh %d did not rotate the token", i+1)
		}
		current = next.RefreshToken
	}

	if live := tokens.liveTokensFor(user.ID); live != 1 {
		t.Errorf("live tokens after a chain of refreshes = %d, want 1", live)
	}
}

// TestRefreshReplayRevokesSessionFamily is the security-critical path: a token
// presented twice means it was copied, so every token for that user dies.
func TestRefreshReplayRevokesSessionFamily(t *testing.T) {
	s, _, tokens := newTestService(t)

	user, pair := registerAndLogin(t, s, "player@example.com", "a good enough password")

	rotated, err := s.Refresh(pair.RefreshToken)
	if err != nil {
		t.Fatalf("first Refresh: %v", err)
	}

	rotationsBefore := tokens.rotations

	// The attacker replays the token the legitimate client already spent.
	if _, err := s.Refresh(pair.RefreshToken); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("replay error = %v, want ErrInvalidRefreshToken", err)
	}
	if tokens.rotations != rotationsBefore {
		t.Error("a replayed token was rotated instead of rejected")
	}

	// The legitimate client's current token is collateral damage, by design:
	// there is no way to tell which side of the exchange was stolen.
	if _, err := s.Refresh(rotated.RefreshToken); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Error("the live token survived a detected replay")
	}
	if live := tokens.liveTokensFor(user.ID); live != 0 {
		t.Errorf("live tokens after a replay = %d, want 0", live)
	}
}

func TestRefreshRejectsUnusableTokens(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		s, _, _ := newTestService(t)
		if _, err := s.Refresh(""); !errors.Is(err, ErrInvalidRefreshToken) {
			t.Errorf("error = %v, want ErrInvalidRefreshToken", err)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		s, _, _ := newTestService(t)
		if _, err := s.Refresh("a-token-nobody-ever-issued"); !errors.Is(err, ErrInvalidRefreshToken) {
			t.Errorf("error = %v, want ErrInvalidRefreshToken", err)
		}
	})

	t.Run("expired", func(t *testing.T) {
		s, _, tokens := newTestService(t)
		_, pair := registerAndLogin(t, s, "player@example.com", "a good enough password")

		stored, err := tokens.GetByHash(auth.HashToken(pair.RefreshToken))
		if err != nil {
			t.Fatalf("stored token not found: %v", err)
		}
		tokens.tokens[stored.ID].ExpiresAt = time.Now().Add(-time.Minute)

		if _, err := s.Refresh(pair.RefreshToken); !errors.Is(err, ErrInvalidRefreshToken) {
			t.Errorf("error = %v, want ErrInvalidRefreshToken", err)
		}
		if tokens.rotations != 0 {
			t.Error("an expired token was rotated")
		}
	})
}

func TestLogoutRevokesToken(t *testing.T) {
	s, _, tokens := newTestService(t)

	user, pair := registerAndLogin(t, s, "player@example.com", "a good enough password")

	if err := s.Logout(pair.RefreshToken); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	if live := tokens.liveTokensFor(user.ID); live != 0 {
		t.Errorf("live tokens after logout = %d, want 0", live)
	}
	if _, err := s.Refresh(pair.RefreshToken); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Error("a logged-out token could still be refreshed")
	}
}

// TestLogoutIsForgiving covers the deliberate leniency: the client's cookies get
// cleared either way, so a token the server cannot place is not an error.
func TestLogoutIsForgiving(t *testing.T) {
	tests := map[string]string{
		"empty token":   "",
		"unknown token": "a-token-nobody-ever-issued",
	}

	for name, token := range tests {
		t.Run(name, func(t *testing.T) {
			s, _, _ := newTestService(t)
			if err := s.Logout(token); err != nil {
				t.Errorf("Logout: %v", err)
			}
		})
	}

	t.Run("already logged out", func(t *testing.T) {
		s, _, _ := newTestService(t)
		_, pair := registerAndLogin(t, s, "player@example.com", "a good enough password")

		if err := s.Logout(pair.RefreshToken); err != nil {
			t.Fatalf("first Logout: %v", err)
		}
		if err := s.Logout(pair.RefreshToken); err != nil {
			t.Errorf("second Logout: %v", err)
		}
	})
}

// TestLogoutLeavesOtherSessionsAlone: signing out of one device must not sign
// the user out everywhere.
func TestLogoutLeavesOtherSessionsAlone(t *testing.T) {
	s, _, tokens := newTestService(t)

	user, phone := registerAndLogin(t, s, "player@example.com", "a good enough password")

	_, laptop, err := s.Login("player@example.com", "a good enough password")
	if err != nil {
		t.Fatalf("second Login: %v", err)
	}
	if live := tokens.liveTokensFor(user.ID); live != 2 {
		t.Fatalf("live tokens across two sessions = %d, want 2", live)
	}

	if err := s.Logout(phone.RefreshToken); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	if live := tokens.liveTokensFor(user.ID); live != 1 {
		t.Errorf("live tokens after one logout = %d, want 1", live)
	}
	if _, err := s.Refresh(laptop.RefreshToken); err != nil {
		t.Errorf("the other session was signed out too: %v", err)
	}
}

func TestGetUser(t *testing.T) {
	s, _, _ := newTestService(t)

	created, err := s.Register(RegisterParams{Email: "player@example.com", Password: "a good enough password"})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	user, err := s.GetUser(created.ID)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if user.Email != "player@example.com" {
		t.Errorf("email = %q, want player@example.com", user.Email)
	}

	// The handler maps this sentinel onto a 401 for a token whose user is gone.
	if _, err := s.GetUser(created.ID + 999); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("error = %v, want gorm.ErrRecordNotFound", err)
	}
}

// errDatabaseDown stands in for any unexpected infrastructure failure.
var errDatabaseDown = errors.New("connection refused")

// TestRegisterSurfacesUnexpectedStoreErrors makes sure an infrastructure
// failure is not quietly reported as a user-facing validation problem.
func TestRegisterSurfacesUnexpectedStoreErrors(t *testing.T) {
	s, users, _ := newTestService(t)
	users.createErr = errDatabaseDown

	_, err := s.Register(RegisterParams{Email: "player@example.com", Password: "a good enough password"})
	if err == nil {
		t.Fatal("Register succeeded despite a failing store")
	}
	if errors.Is(err, ErrEmailTaken) {
		t.Error("a store failure was reported as a taken email")
	}
	if !errors.Is(err, errDatabaseDown) {
		t.Errorf("error = %v, does not wrap the underlying failure", err)
	}
}

// TestLoginSurfacesUnexpectedStoreErrors: a lookup that fails for infrastructure
// reasons must not be flattened into "wrong password", which would send the user
// round in circles retyping correct credentials.
func TestLoginSurfacesUnexpectedStoreErrors(t *testing.T) {
	t.Run("user lookup fails", func(t *testing.T) {
		s, users, _ := newTestService(t)
		users.lookupErr = errDatabaseDown

		_, _, err := s.Login("player@example.com", "a good enough password")
		if errors.Is(err, ErrInvalidCredentials) {
			t.Error("a store failure was reported as invalid credentials")
		}
		if !errors.Is(err, errDatabaseDown) {
			t.Errorf("error = %v, does not wrap the underlying failure", err)
		}
	})

	t.Run("storing the refresh token fails", func(t *testing.T) {
		s, _, tokens := newTestService(t)
		if _, err := s.Register(RegisterParams{Email: "player@example.com", Password: "a good enough password"}); err != nil {
			t.Fatalf("Register: %v", err)
		}
		tokens.createErr = errDatabaseDown

		_, pair, err := s.Login("player@example.com", "a good enough password")
		if !errors.Is(err, errDatabaseDown) {
			t.Errorf("error = %v, does not wrap the underlying failure", err)
		}
		// A session the server cannot refresh is worse than no session.
		if pair.AccessToken != "" {
			t.Error("an access token was handed out without a stored refresh token")
		}
	})
}

// TestRefreshSurfacesUnexpectedStoreErrors: the refresh endpoint clears cookies
// on ErrInvalidRefreshToken, so a database blip reported as one would sign every
// active user out.
func TestRefreshSurfacesUnexpectedStoreErrors(t *testing.T) {
	t.Run("token lookup fails", func(t *testing.T) {
		s, _, tokens := newTestService(t)
		_, pair := registerAndLogin(t, s, "player@example.com", "a good enough password")
		tokens.lookupErr = errDatabaseDown

		_, err := s.Refresh(pair.RefreshToken)
		if errors.Is(err, ErrInvalidRefreshToken) {
			t.Error("a store failure was reported as an invalid refresh token")
		}
		if !errors.Is(err, errDatabaseDown) {
			t.Errorf("error = %v, does not wrap the underlying failure", err)
		}
	})

	t.Run("rotation fails", func(t *testing.T) {
		s, _, tokens := newTestService(t)
		_, pair := registerAndLogin(t, s, "player@example.com", "a good enough password")
		tokens.rotateErr = errDatabaseDown

		if _, err := s.Refresh(pair.RefreshToken); !errors.Is(err, errDatabaseDown) {
			t.Errorf("error = %v, does not wrap the underlying failure", err)
		}
	})
}

// TestRefreshLosesRotationRace covers two requests refreshing the same token at
// once: the repository reports the loser as not-found, and it must come back as
// a spent token rather than a server error.
func TestRefreshLosesRotationRace(t *testing.T) {
	s, _, tokens := newTestService(t)
	_, pair := registerAndLogin(t, s, "player@example.com", "a good enough password")

	tokens.rotateErr = gorm.ErrRecordNotFound

	if _, err := s.Refresh(pair.RefreshToken); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("error = %v, want ErrInvalidRefreshToken", err)
	}
}

// TestLogoutSurfacesUnexpectedStoreErrors: leniency covers tokens the server
// cannot place, not a store it cannot reach.
func TestLogoutSurfacesUnexpectedStoreErrors(t *testing.T) {
	s, _, tokens := newTestService(t)
	_, pair := registerAndLogin(t, s, "player@example.com", "a good enough password")
	tokens.lookupErr = errDatabaseDown

	if err := s.Logout(pair.RefreshToken); !errors.Is(err, errDatabaseDown) {
		t.Errorf("error = %v, does not wrap the underlying failure", err)
	}
}
