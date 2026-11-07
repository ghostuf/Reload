package repository

import (
	"github.com/ghostuf/reload/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a user. A duplicate email surfaces as gorm.ErrDuplicatedKey.
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// GetByEmail looks up a user by exact email. Callers are expected to have
// normalised the address first.
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetRole returns a user's role without loading the rest of the row. AdminOnly
// calls this on every admin request, so it stays a single narrow column read.
func (r *UserRepository) GetRole(id uint) (models.UserRole, error) {
	var user models.User
	if err := r.db.Select("role").First(&user, id).Error; err != nil {
		return "", err
	}
	return user.Role, nil
}

// UpdateRole sets a user's role directly, bypassing Save so it never touches
// the password hash or any other column.
func (r *UserRepository) UpdateRole(id uint, role models.UserRole) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("role", role).Error
}
