package handlers

import (
	"errors"
	"net/http"

	"github.com/ghostuf/reload/internal/auth"
	"github.com/ghostuf/reload/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EsewaHandler struct {
	service     *services.EsewaCheckoutService
	paymentMode string
}

func NewEsewaHandler(service *services.EsewaCheckoutService, paymentMode string) *EsewaHandler {
	return &EsewaHandler{service: service, paymentMode: paymentMode}
}

// Initiate prices the cart server-side, creates a pending order and
// payment, and returns the signed form fields the frontend must POST to
// eSewa's own form URL.
func (h *EsewaHandler) Initiate(c *gin.Context) {
	if h.paymentMode != "esewa" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "esewa payment mode is not active"})
		return
	}

	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	result, err := h.service.InitiatePayment(userID)
	if err != nil {
		respondEsewaError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": gin.H{
		"order_id": result.OrderID,
		"form":     result.Form,
		"form_url": result.FormURL,
	}})
}

type esewaVerifyRequest struct {
	Data string `json:"data"`
}

// Verify checks the signature and independent status of the payment
// eSewa's success redirect names, and completes it exactly once.
func (h *EsewaHandler) Verify(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var req esewaVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Data == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	order, err := h.service.VerifyPayment(userID, req.Data)
	if err != nil {
		respondEsewaError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": order})
}

type esewaFailureRequest struct {
	TransactionUUID string `json:"transaction_uuid"`
}

// Failure marks a pending payment cancelled or failed, leaving the cart
// untouched so the user can retry from scratch.
func (h *EsewaHandler) Failure(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var req esewaFailureRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.TransactionUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.FailPayment(userID, req.TransactionUUID); err != nil {
		respondEsewaError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"cancelled": true}})
}

// respondEsewaError maps service-level sentinel errors onto status codes,
// and keeps unexpected failures from leaking internals to the client.
func respondEsewaError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrCartEmpty), errors.Is(err, services.ErrNothingToPurchase):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrEsewaInvalidPayload), errors.Is(err, services.ErrEsewaSignatureMismatch):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrEsewaAmountMismatch), errors.Is(err, services.ErrEsewaPaymentNotComplete):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "payment request failed"})
	}
}
