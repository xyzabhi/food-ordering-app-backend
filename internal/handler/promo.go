package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xyzabhi/food-ordering-app-backend/internal/coupons"
)

type PromoHandler struct {
	store       coupons.Store
	discountPct float64
}

func NewPromoHandler(store coupons.Store, discountPct float64) *PromoHandler {
	return &PromoHandler{store: store, discountPct: discountPct}
}

// CheckPromo validates a promo code.
// GET /checkpromo?code=ABC123...
func (h *PromoHandler) CheckPromo(c *gin.Context) {
	code := strings.TrimSpace(c.Query("code"))
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing promo code"})
		return
	}

	valid := h.store != nil && h.store.IsValid(code)
	c.JSON(http.StatusOK, gin.H{
		"code":            code,
		"discountPercent": h.discountPct,
		"valid":           valid,
	})
}

