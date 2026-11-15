package repository

import (
	"fmt"

	"github.com/ghostuf/reload/internal/models"
	"gorm.io/gorm"
)

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create inserts a payment. Called against a transaction-scoped repository
// alongside the order it belongs to, so the two land together.
func (r *PaymentRepository) Create(payment *models.Payment) error {
	return r.db.Create(payment).Error
}

// GetByTransactionUUID returns a payment with its order -- and that order's
// items and each item's game -- preloaded, so a caller has everything
// needed to check ownership and build a receipt without a second query.
// The game preload is Unscoped for the same reason OrderRepository.GetByID's
// is: a soft-deleted game must not vanish from a receipt that already
// references it.
func (r *PaymentRepository) GetByTransactionUUID(transactionUUID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.
		Preload("Order.Items.Game", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		}).
		Where("transaction_uuid = ?", transactionUUID).
		First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// MarkCompleted records a successful payment -- its provider reference, raw
// response, and completed status -- in one update.
func (r *PaymentRepository) MarkCompleted(id uint, providerRefID *string, rawResponse models.RawJSON) error {
	err := r.db.Model(&models.Payment{}).Where("id = ?", id).Updates(map[string]any{
		"status":          models.PaymentStatusCompleted,
		"provider_ref_id": providerRefID,
		"raw_response":    rawResponse,
	}).Error
	if err != nil {
		return fmt.Errorf("marking payment completed: %w", err)
	}
	return nil
}

// MarkStatus sets a payment's status alone, e.g. to failed or cancelled.
func (r *PaymentRepository) MarkStatus(id uint, status models.PaymentStatus) error {
	if err := r.db.Model(&models.Payment{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return fmt.Errorf("updating payment status: %w", err)
	}
	return nil
}
