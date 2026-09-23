package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/ghostuf/reload/internal/auth"
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
// history list without shipping every line item's full game details.
type orderSummary struct {
	ID         uint      `json:"id"`
	TotalPrice float64   `json:"total_price"`
	ItemCount  int       `json:"item_count"`
	CreatedAt  time.Time `json:"created_at"`
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
		summaries[i] = orderSummary{
			ID:         order.ID,
			TotalPrice: order.TotalPrice,
			ItemCount:  len(order.Items),
			CreatedAt:  order.CreatedAt,
		}
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
