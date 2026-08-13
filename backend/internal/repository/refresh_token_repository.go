package repository

import (
	"fmt"

	"github.com/ghostuf/reload/internal/models"
	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(token *models.RefreshToken) error {
	return r.db.Create(token).Error
}

// GetByHash finds a stored refresh token by the hash of its raw value. It
// returns gorm.ErrRecordNotFound when no such token exists.
func (r *RefreshTokenRepository) GetByHash(hash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	if err := r.db.Where("token_hash = ?", hash).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

// Revoke marks a single token as revoked. Revoking an already-revoked token is
// a no-op, which keeps logout idempotent.
func (r *RefreshTokenRepository) Revoke(id uint) error {
	err := r.db.Model(&models.RefreshToken{}).
		Where("id = ?", id).
		Update("revoked", true).Error
	if err != nil {
		return fmt.Errorf("revoking refresh token %d: %w", id, err)
	}
	return nil
}

// RevokeAllForUser revokes every live token belonging to a user. Used to shut
// down a session family when a revoked token is replayed.
func (r *RefreshTokenRepository) RevokeAllForUser(userID uint) error {
	err := r.db.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Update("revoked", true).Error
	if err != nil {
		return fmt.Errorf("revoking refresh tokens for user %d: %w", userID, err)
	}
	return nil
}

// Rotate atomically revokes the current token and stores its replacement, so a
// crash can never leave a session with two live tokens or none.
func (r *RefreshTokenRepository) Rotate(oldID uint, replacement *models.RefreshToken) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.RefreshToken{}).
			Where("id = ? AND revoked = ?", oldID, false).
			Update("revoked", true)
		if result.Error != nil {
			return result.Error
		}
		// Another request rotated this same token first; abort rather than
		// issuing a second live token for the session.
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return tx.Create(replacement).Error
	})
	if err != nil {
		return fmt.Errorf("rotating refresh token %d: %w", oldID, err)
	}
	return nil
}
