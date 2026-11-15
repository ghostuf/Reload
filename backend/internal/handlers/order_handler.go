package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/ghostuf/reload/internal/auth"
	"github.com/ghostuf/reload/internal/models"
	"github.com/ghostuf/reload/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderHandler struct {
	service *services.OrderService
}

func NewOrderHandler(service *services.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

// orderSummary is what GET /orders returns per order: enough to render a
// history list -- including payment status and eSewa's own reference code,
// when there is a payment at all -- without shipping every line item's
// full game details.
type orderSummary struct {
	ID              uint      `json:"id"`
	TotalPrice      float64   `json:"total_price"`
	ItemCount       int       `json:"item_count"`
	Status          string    `json:"status"`
	PaymentProvider *string   `json:"payment_provider,omitempty"`
	PaymentStatus   *string   `json:"payment_status,omitempty"`
	ProviderRefID   *string   `json:"provider_ref_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// toOrderSummary projects an order (and its payment, if the mock checkout
// didn't skip creating one) onto the lighter shape the history list needs.
func toOrderSummary(order models.Order) orderSummary {
	summary := orderSummary{
		ID:         order.ID,
		TotalPrice: order.TotalPrice,
		ItemCount:  len(order.Items),
		Status:     string(order.Status),
		CreatedAt:  order.CreatedAt,
	}

	if order.Payment != nil {
		provider := string(order.Payment.Provider)
		status := string(order.Payment.Status)
		summary.PaymentProvider = &provider
		summary.PaymentStatus = &status
		summary.ProviderRefID = order.Payment.ProviderRefID
	}

	return summary
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	orders, err := h.service.ListOrders(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load orders"})
		return
	}

	summaries := make([]orderSummary, len(orders))
	for i, order := range orders {
		summaries[i] = toOrderSummary(order)
	}

	c.JSON(http.StatusOK, gin.H{"data": summaries})
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	order, err := h.service.GetOrder(userID, uint(orderID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": order})
}
