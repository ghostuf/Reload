package services

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/ghostuf/reload/internal/models"
	"github.com/ghostuf/reload/internal/repository"
	"gorm.io/gorm"
)

var (
	// ErrEsewaAmountMismatch is returned when eSewa's own status-check
	// response reports a different amount than the order it is supposed to
	// be paying for -- the fraud case eSewa's own documentation calls out
	// by name.
	ErrEsewaAmountMismatch = errors.New("payment amount does not match the order total")
	// ErrEsewaPaymentNotComplete is returned when eSewa's status-check
	// response is anything other than COMPLETE.
	ErrEsewaPaymentNotComplete = errors.New("payment is not complete")
)

// staleOrderAge is how long a pending order is given to actually get paid
// before OrderRepository.ExpireStalePending gives up on it. eSewa's own
// login session expires after about 15 minutes; this is a deliberately
// generous multiple of that so a slow but legitimate payment never gets
// cancelled out from under the user.
const staleOrderAge = 30 * time.Minute

// EsewaConfig is the subset of application configuration
// EsewaCheckoutService needs. SecretKey never appears in anything this
// service returns to a handler -- only the signature computed from it does.
type EsewaConfig struct {
	SecretKey   string
	ProductCode string
	FrontendURL string
	// FormURL is eSewa's own form-submission URL, returned to the frontend
	// alongside the signed fields so it knows where to POST them -- the
	// frontend has no env var of its own for this; the backend is the one
	// place ESEWA_ENV already decides sandbox vs production.
	FormURL string
}

// EsewaPaymentFinder is the read access VerifyPayment and FailPayment need
// to look up a payment by its transaction_uuid, with its order attached.
// Pulled out as its own interface -- rather than going through s.db
// directly, the way the rest of this service does for its writes -- so the
// signature check, ownership check, and idempotent short-circuit in
// VerifyPayment can be exercised without a database at all.
type EsewaPaymentFinder interface {
	GetByTransactionUUID(transactionUUID string) (*models.Payment, error)
}

var _ EsewaPaymentFinder = (*repository.PaymentRepository)(nil)

// EsewaCheckoutService holds the raw *gorm.DB for its writes, for the same
// reason CheckoutService does: initiating and completing a payment each
// read and write across cart, order, payment, and library tables inside
// one transaction. Lookups go through EsewaPaymentFinder instead.
type EsewaCheckoutService struct {
	db       *gorm.DB
	payments EsewaPaymentFinder
	status   EsewaStatusChecker
	cfg      EsewaConfig
}

func NewEsewaCheckoutService(db *gorm.DB, payments EsewaPaymentFinder, status EsewaStatusChecker, cfg EsewaConfig) *EsewaCheckoutService {
	return &EsewaCheckoutService{db: db, payments: payments, status: status, cfg: cfg}
}

// EsewaInitiateResult is what POST /checkout/esewa/initiate returns: the id
// of the pending order just created, and every field the frontend must POST
// to eSewa's form URL to actually pay for it.
type EsewaInitiateResult struct {
	OrderID uint
	Form    EsewaFormFields
	FormURL string
}

// InitiatePayment prices the user's cart the same way the mock checkout
// does -- same exclusion of already-owned games, same effective-price
// helper, server-computed total only -- and creates a pending order and
// payment for it. It does not touch the library or the cart: those only
// happen once VerifyPayment confirms the money actually moved.
func (s *EsewaCheckoutService) InitiatePayment(userID uint) (EsewaInitiateResult, error) {
	var (
		orderID         uint
		transactionUUID string
		total           float64
	)

	err := s.db.Transaction(func(tx *gorm.DB) error {
		cartRepo := repository.NewCartRepository(tx)
		libraryRepo := repository.NewLibraryRepository(tx)
		orderRepo := repository.NewOrderRepository(tx)
		paymentRepo := repository.NewPaymentRepository(tx)

		cartItems, err := cartRepo.ListByUser(userID)
		if err != nil {
			return fmt.Errorf("loading cart: %w", err)
		}
		if len(cartItems) == 0 {
			return ErrCartEmpty
		}

		owned, err := libraryRepo.OwnedGameIDs(userID)
		if err != nil {
			return fmt.Errorf("loading library: %w", err)
		}

		orderItems, orderTotal := buildOrderItems(cartItems, owned)
		if len(orderItems) == 0 {
			return ErrNothingToPurchase
		}
		total = orderTotal

		order := models.Order{
			UserID:     userID,
			TotalPrice: total,
			Status:     models.OrderStatusPending,
			Items:      orderItems,
		}
		if err := orderRepo.Create(&order); err != nil {
			return fmt.Errorf("creating order: %w", err)
		}
		orderID = order.ID

		uuid, err := generateTransactionUUID()
		if err != nil {
			return err
		}
		transactionUUID = uuid

		payment := models.Payment{
			OrderID:         order.ID,
			Provider:        models.PaymentProviderEsewa,
			TransactionUUID: transactionUUID,
			Amount:          total,
			Status:          models.PaymentStatusPending,
		}
		if err := paymentRepo.Create(&payment); err != nil {
			return fmt.Errorf("creating payment: %w", err)
		}

		return nil
	})
	if err != nil {
		return EsewaInitiateResult{}, err
	}

	// eSewa's success_url gets a ?data= payload identifying the transaction;
	// its documentation does not promise the same for failure_url, so the
	// transaction_uuid is embedded in the failure URL's own query string --
	// ours to set, and eSewa redirects back to it verbatim -- so the
	// frontend's failure page always has something to report against
	// regardless of what, if anything, eSewa appends to it.
	form := buildEsewaFormFields(
		s.cfg.SecretKey,
		s.cfg.ProductCode,
		transactionUUID,
		total,
		s.cfg.FrontendURL+"/checkout/esewa/success",
		s.cfg.FrontendURL+"/checkout/esewa/failure?transaction_uuid="+url.QueryEscape(transactionUUID),
	)

	return EsewaInitiateResult{OrderID: orderID, Form: form, FormURL: s.cfg.FormURL}, nil
}

// VerifyPayment decodes and checks the signature on eSewa's success
// callback, confirms the payment independently through eSewa's own
// status-check API, and -- only once both agree it actually completed, for
// the right amount -- grants the library, clears the cart, and marks the
// order and payment completed, all in one transaction.
//
// It is idempotent: if the order is already completed, it is returned as
// it stands rather than redoing any of that, since the user is free to
// refresh the success page, or land on it twice from a slow redirect,
// without anything being charged or granted twice.
func (s *EsewaCheckoutService) VerifyPayment(userID uint, data string) (*models.Order, error) {
	fields, err := decodeEsewaCallback(data)
	if err != nil {
		return nil, err
	}

	if !verifyEsewaSignature(s.cfg.SecretKey, fields) {
		return nil, ErrEsewaSignatureMismatch
	}

	transactionUUID := fields["transaction_uuid"]

	payment, err := s.payments.GetByTransactionUUID(transactionUUID)
	if err != nil {
		return nil, err
	}

	// Scoped to the authenticated user the same way every other order
	// lookup in this app is: a payment that exists but belongs to someone
	// else reads as not-found rather than forbidden, so this can never be
	// used to probe for valid transaction_uuids.
	if payment.Order.UserID != userID {
		return nil, gorm.ErrRecordNotFound
	}

	// Idempotent: a refreshed success page re-sends the same signed
	// callback and must not re-run any of the side effects below.
	if payment.Order.Status == models.OrderStatusCompleted {
		return &payment.Order, nil
	}

	result, err := s.status.CheckStatus(s.cfg.ProductCode, transactionUUID, payment.Amount)
	if err != nil {
		return nil, fmt.Errorf("checking payment status: %w", err)
	}

	if result.Status != EsewaStatusComplete {
		return nil, fmt.Errorf("%w: eSewa reports %q", ErrEsewaPaymentNotComplete, result.Status)
	}

	// eSewa's own FAQ calls this out directly as the fraud case: a COMPLETE
	// status for a different amount than what was actually requested must
	// stop the transaction rather than honour it.
	if !amountsMatch(result.TotalAmount, payment.Amount) {
		_ = s.markFailed(payment.ID, payment.OrderID)
		return nil, ErrEsewaAmountMismatch
	}

	return s.completePayment(userID, payment, result)
}

// FailPayment marks a pending order and payment as cancelled, leaving the
// cart untouched so the user can retry. An order that is not pending --
// most likely already completed, if a failure notice arrives after the
// payment somehow already succeeded -- is left alone rather than
// downgraded.
func (s *EsewaCheckoutService) FailPayment(userID uint, transactionUUID string) error {
	payment, err := s.payments.GetByTransactionUUID(transactionUUID)
	if err != nil {
		return err
	}
	if payment.Order.UserID != userID {
		return gorm.ErrRecordNotFound
	}
	if payment.Order.Status != models.OrderStatusPending {
		return nil
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := repository.NewPaymentRepository(tx).MarkStatus(payment.ID, models.PaymentStatusCancelled); err != nil {
			return err
		}
		return repository.NewOrderRepository(tx).MarkStatus(payment.OrderID, models.OrderStatusCancelled)
	})
}

// completePayment marks the payment and order completed, grants the
// library, and clears the cart, all inside one transaction: a failure
// partway through must not leave a completed payment with no library
// grant, or a cleared cart with an order still sitting pending.
func (s *EsewaCheckoutService) completePayment(userID uint, payment *models.Payment, result EsewaStatusResult) (*models.Order, error) {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := repository.NewPaymentRepository(tx).MarkCompleted(payment.ID, result.RefID, models.RawJSON(result.Raw)); err != nil {
			return err
		}
		if err := repository.NewOrderRepository(tx).MarkStatus(payment.OrderID, models.OrderStatusCompleted); err != nil {
			return err
		}

		acquiredAt := time.Now()
		libraryItems := make([]models.LibraryItem, len(payment.Order.Items))
		for i, item := range payment.Order.Items {
			libraryItems[i] = models.LibraryItem{
				UserID:     userID,
				GameID:     item.GameID,
				AcquiredAt: acquiredAt,
			}
		}
		if err := repository.NewLibraryRepository(tx).CreateManyIgnoringConflicts(libraryItems); err != nil {
			return err
		}

		if err := repository.NewCartRepository(tx).ClearByUser(userID); err != nil {
			return fmt.Errorf("clearing cart: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return repository.NewOrderRepository(s.db).GetByID(userID, payment.OrderID)
}

// markFailed marks a payment and its order failed. Used from an error path
// that is already returning a more specific error to its caller, so a
// secondary failure here is deliberately swallowed rather than compounding
// it.
func (s *EsewaCheckoutService) markFailed(paymentID, orderID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := repository.NewPaymentRepository(tx).MarkStatus(paymentID, models.PaymentStatusFailed); err != nil {
			return err
		}
		return repository.NewOrderRepository(tx).MarkStatus(orderID, models.OrderStatusFailed)
	})
}

// amountsMatch compares two amounts allowing a cent of float rounding
// slack, rather than demanding bit-for-bit equality between a value this
// app computed and one eSewa's own API reported back.
func amountsMatch(a, b float64) bool {
	const epsilon = 0.01
	diff := a - b
	return diff > -epsilon && diff < epsilon
}
