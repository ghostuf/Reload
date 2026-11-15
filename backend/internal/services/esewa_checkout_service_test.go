package services

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ghostuf/reload/internal/models"
)

// fakeEsewaPaymentFinder is an in-memory EsewaPaymentFinder.
type fakeEsewaPaymentFinder struct {
	payments map[string]*models.Payment
}

func (f *fakeEsewaPaymentFinder) GetByTransactionUUID(transactionUUID string) (*models.Payment, error) {
	payment, ok := f.payments[transactionUUID]
	if !ok {
		return nil, errors.New("payment not found")
	}
	return payment, nil
}

// fakeEsewaStatusChecker records whether it was called at all, so a test
// can assert the idempotent path never reaches it -- an unnecessary call to
// eSewa's status-check API on every page refresh would be its own kind of
// bug, even if a harmless one.
type fakeEsewaStatusChecker struct {
	called bool
	result EsewaStatusResult
	err    error
}

func (f *fakeEsewaStatusChecker) CheckStatus(productCode, transactionUUID string, totalAmount float64) (EsewaStatusResult, error) {
	f.called = true
	return f.result, f.err
}

// signedEsewaCallback builds a success-callback "data" payload for
// transactionUUID/totalAmount, signed the same way eSewa signs a real one,
// base64-encoded exactly as it would arrive on success_url.
func signedEsewaCallback(t *testing.T, secret, transactionUUID, totalAmount string) string {
	t.Helper()

	fields := map[string]string{
		"transaction_code": "0LD5CEH",
		"status":           "COMPLETE",
		"total_amount":     totalAmount,
		"transaction_uuid": transactionUUID,
		"product_code":     "EPAYTEST",
	}
	signedFieldNames := []string{"transaction_code", "status", "total_amount", "transaction_uuid", "product_code", "signed_field_names"}
	fields["signed_field_names"] = "transaction_code,status,total_amount,transaction_uuid,product_code,signed_field_names"
	fields["signature"] = esewaSign(secret, esewaSignatureMessage(signedFieldNames, fields))

	raw, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("marshalling callback fields: %v", err)
	}
	return base64.StdEncoding.EncodeToString(raw)
}

// TestVerifyPaymentIsIdempotent is the case the user hitting refresh (or a
// slow redirect landing twice) depends on: a second call for an
// already-completed order must return it as-is, without calling eSewa's
// status-check API again and without touching the database a second time.
// The service is constructed with a nil *gorm.DB specifically so that any
// attempt to write through it would panic -- proving this path truly never
// gets there, not just that it happens not to today.
func TestVerifyPaymentIsIdempotent(t *testing.T) {
	const (
		userID          = uint(7)
		transactionUUID = "240613-134231"
	)

	completedOrder := models.Order{
		ID:         5,
		UserID:     userID,
		Status:     models.OrderStatusCompleted,
		TotalPrice: 1000,
	}
	finder := &fakeEsewaPaymentFinder{
		payments: map[string]*models.Payment{
			transactionUUID: {
				ID:              1,
				OrderID:         5,
				TransactionUUID: transactionUUID,
				Amount:          1000,
				Status:          models.PaymentStatusCompleted,
				Order:           completedOrder,
			},
		},
	}
	statusChecker := &fakeEsewaStatusChecker{}

	svc := NewEsewaCheckoutService(nil, finder, statusChecker, EsewaConfig{
		SecretKey:   testEsewaSecret,
		ProductCode: "EPAYTEST",
	})

	data := signedEsewaCallback(t, testEsewaSecret, transactionUUID, "1,000.0")

	order, err := svc.VerifyPayment(userID, data)
	if err != nil {
		t.Fatalf("VerifyPayment: %v", err)
	}
	if order.Status != models.OrderStatusCompleted {
		t.Errorf("order status = %q, want completed", order.Status)
	}
	if statusChecker.called {
		t.Error("an already-completed order re-checked status with eSewa instead of short-circuiting")
	}

	// A second refresh must behave identically.
	order2, err := svc.VerifyPayment(userID, data)
	if err != nil {
		t.Fatalf("second VerifyPayment: %v", err)
	}
	if order2.ID != order.ID {
		t.Errorf("second call returned a different order: %d vs %d", order2.ID, order.ID)
	}
	if statusChecker.called {
		t.Error("eSewa's status-check API was called despite the order already being completed")
	}
}

// TestVerifyPaymentRejectsWrongUser covers the ownership check: a payment
// that exists but belongs to someone else must read as not-found, the same
// way every other order lookup in this app does, rather than leaking that
// the transaction_uuid is valid.
func TestVerifyPaymentRejectsWrongUser(t *testing.T) {
	const transactionUUID = "240613-134231"

	finder := &fakeEsewaPaymentFinder{
		payments: map[string]*models.Payment{
			transactionUUID: {
				ID:      1,
				OrderID: 5,
				Amount:  1000,
				Status:  models.PaymentStatusPending,
				Order:   models.Order{ID: 5, UserID: 99, Status: models.OrderStatusPending},
			},
		},
	}
	statusChecker := &fakeEsewaStatusChecker{}

	svc := NewEsewaCheckoutService(nil, finder, statusChecker, EsewaConfig{
		SecretKey:   testEsewaSecret,
		ProductCode: "EPAYTEST",
	})

	data := signedEsewaCallback(t, testEsewaSecret, transactionUUID, "1,000.0")

	// userID 7 does not own the order, which belongs to user 99.
	if _, err := svc.VerifyPayment(7, data); err == nil {
		t.Error("VerifyPayment succeeded for a payment belonging to a different user")
	}
	if statusChecker.called {
		t.Error("eSewa's status-check API was called before the ownership check even passed")
	}
}

// TestVerifyPaymentRejectsBadSignature covers the other thing that must
// never reach the database or eSewa's status-check API: a response whose
// signature does not match.
func TestVerifyPaymentRejectsBadSignature(t *testing.T) {
	const transactionUUID = "240613-134231"

	finder := &fakeEsewaPaymentFinder{
		payments: map[string]*models.Payment{
			transactionUUID: {
				ID:      1,
				OrderID: 5,
				Amount:  1000,
				Status:  models.PaymentStatusPending,
				Order:   models.Order{ID: 5, UserID: 7, Status: models.OrderStatusPending},
			},
		},
	}
	statusChecker := &fakeEsewaStatusChecker{}

	svc := NewEsewaCheckoutService(nil, finder, statusChecker, EsewaConfig{
		SecretKey:   testEsewaSecret,
		ProductCode: "EPAYTEST",
	})

	// Signed with the wrong secret, so the signature cannot possibly match.
	data := signedEsewaCallback(t, "a-different-secret", transactionUUID, "1,000.0")

	if _, err := svc.VerifyPayment(7, data); !errors.Is(err, ErrEsewaSignatureMismatch) {
		t.Errorf("error = %v, want ErrEsewaSignatureMismatch", err)
	}
	if statusChecker.called {
		t.Error("eSewa's status-check API was called despite a bad signature")
	}
}
