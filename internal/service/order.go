package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/xyzabhi/food-ordering-app-backend/internal/dto"
	"github.com/xyzabhi/food-ordering-app-backend/internal/repository"
)

var (
	ErrUnknownProduct = errors.New("unknown product id")
	ErrInvalidCoupon  = errors.New("invalid coupon code")
)

// couponCode -> discount percent (0–100) on subtotal after line totals
var couponPercentOff = map[string]float64{
	"SAVE10": 10,
	"SAVE20": 20,
}

type OrderService struct {
	orderRepository   repository.OrderRepository
	productRepository repository.ProductRepository
}

func NewOrderService(orderRepo repository.OrderRepository, productRepo repository.ProductRepository) *OrderService {
	return &OrderService{
		orderRepository:   orderRepo,
		productRepository: productRepo,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, req dto.OrderRequest) (dto.OrderResponse, error) {
	if len(req.Items) == 0 {
		return dto.OrderResponse{}, errors.New("order has no items")
	}

	merged := make(map[string]int)
	for _, it := range req.Items {
		merged[it.ProductID] += it.Quantity
	}

	ids := make([]string, 0, len(merged))
	for id := range merged {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	products, err := s.productRepository.GetByIDs(ctx, ids)
	if err != nil {
		return dto.OrderResponse{}, err
	}

	var total float64
	responseItems := make([]dto.OrderItemRequest, 0, len(merged))
	productResponses := make([]dto.ProductResponse, 0, len(merged))
	lines := make([]repository.OrderLine, 0, len(merged))

	for _, productID := range ids {
		qty := merged[productID]
		p, ok := products[productID]
		if !ok {
			return dto.OrderResponse{}, fmt.Errorf("%w: %s", ErrUnknownProduct, productID)
		}
		lineTotal := p.Price * float64(qty)
		total += lineTotal
		responseItems = append(responseItems, dto.OrderItemRequest{
			ProductID: productID,
			Quantity:  qty,
		})
		productResponses = append(productResponses, dto.ProductResponse{
			ID:       p.ID,
			Name:     p.Name,
			Price:    p.Price,
			Category: p.Category,
			Image:    p.Image,
		})
		lines = append(lines, repository.OrderLine{ProductID: productID, Quantity: qty})
	}

	coupon := strings.TrimSpace(req.CouponCode)
	var discount float64
	if coupon != "" {
		pct, ok := couponPercentOff[strings.ToUpper(coupon)]
		if !ok {
			return dto.OrderResponse{}, ErrInvalidCoupon
		}
		discount = total * (pct / 100)
	}

	final := total - discount
	if final < 0 {
		final = 0
	}

	orderID := uuid.NewString()
	createdAt, err := s.orderRepository.CreateOrder(ctx, repository.OrderCreateInput{
		ID:         orderID,
		CouponCode: coupon,
		TotalPrice: total,
		Discount:   discount,
		FinalPrice: final,
		Lines:      lines,
	})
	if err != nil {
		return dto.OrderResponse{}, err
	}

	return dto.OrderResponse{
		ID:         orderID,
		CouponCode: coupon,
		Items:      responseItems,
		Products:   productResponses,
		TotalPrice: total,
		Discount:   discount,
		FinalPrice: final,
		CreatedAt:  createdAt,
	}, nil
}
