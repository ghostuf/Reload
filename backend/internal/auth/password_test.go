package auth

import (
	"errors"
	"strings"
	"testing"
)

func TestHashPasswordRoundTrip(t *testing.T) {
	const password = "correct horse battery staple"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if hash == password {
		t.Fatal("hash equals the plaintext password")
	}
	if strings.Contains(hash, password) {
		t.Fatal("hash contains the plaintext password")
	}

	if err := VerifyPassword(hash, password); err != nil {
		t.Errorf("VerifyPassword with the correct password: %v", err)
	}
}

func TestHashPasswordIsSalted(t *testing.T) {
	const password = "the same password twice"

	first, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	second, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	// Equal digests would mean an unsalted hash, letting one rainbow table
	// crack every account that shares a password.
	if first == second {
		t.Error("hashing the same password twice produced identical hashes")
	}
	if err := VerifyPassword(second, password); err != nil {
		t.Errorf("VerifyPassword against the second hash: %v", err)
	}
}

func TestVerifyPasswordRejectsWrongPassword(t *testing.T) {
	hash, err := HashPassword("the real password")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	for name, attempt := range map[string]string{
		"wrong password": "not the real password",
		"empty":          "",
		"prefix only":    "the real",
		"wrong case":     "The Real Password",
	} {
		t.Run(name, func(t *testing.T) {
			if err := VerifyPassword(hash, attempt); !errors.Is(err, ErrPasswordMismatch) {
				t.Errorf("error = %v, want ErrPasswordMismatch", err)
			}
		})
	}
}

// TestVerifyPasswordRejectsMalformedHash covers a corrupted or empty stored
// hash. That is a data problem, not a wrong password, so it must not surface as
// ErrPasswordMismatch — otherwise it would read as an ordinary failed login.
func TestVerifyPasswordRejectsMalformedHash(t *testing.T) {
	for name, hash := range map[string]string{
		"empty":     "",
		"plaintext": "hunter2",
		"truncated": "$2a$10$N9qo8uLOickgx2ZMRZoMye",
	} {
		t.Run(name, func(t *testing.T) {
			err := VerifyPassword(hash, "any password")
			if err == nil {
				t.Fatal("malformed hash verified successfully")
			}
			if errors.Is(err, ErrPasswordMismatch) {
				t.Error("malformed hash reported as a password mismatch")
			}
		})
	}
}

// TestVerifyPasswordAtBcryptLimit documents the 72-byte ceiling the service
// enforces: bcrypt ignores bytes past it, so two passwords sharing a 72-byte
// prefix verify against the same hash.
func TestVerifyPasswordAtBcryptLimit(t *testing.T) {
	limit := strings.Repeat("a", 72)

	hash, err := HashPassword(limit)
	if err != nil {
		t.Fatalf("HashPassword at 72 bytes: %v", err)
	}
	if err := VerifyPassword(hash, limit); err != nil {
		t.Errorf("VerifyPassword at 72 bytes: %v", err)
	}
	// Truncation is the reason for the cap: without it, a user typing a long
	// passphrase would be authenticated by any string sharing its first 72
	// bytes, with nothing in the API hinting that the tail was discarded.
	if err := VerifyPassword(hash, limit+"ignored"); err != nil {
		t.Errorf("bcrypt no longer truncates at 72 bytes (%v); revisit maxPasswordLength", err)
	}
}
