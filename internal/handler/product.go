package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xyzabhi/food-ordering-app-backend/internal/dto"
	"github.com/xyzabhi/food-ordering-app-backend/internal/repository"
)

type ProductHandler struct {
	productRepo repository.ProductRepository
}

func NewProductHandler(productRepo repository.ProductRepository) *ProductHandler {
	return &ProductHandler{productRepo: productRepo}
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
	products, err := h.productRepo.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]dto.ProductResponse, 0, len(products))
	for _, p := range products {
		out = append(out, dto.ProductResponse{
			ID:       p.ID,
			Name:     p.Name,
			Price:    p.Price,
			Category: p.Category,
			Image:    p.Image,
		})
	}

	c.JSON(http.StatusOK, gin.H{"products": out})
}
