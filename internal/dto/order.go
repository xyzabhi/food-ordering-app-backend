package dto

import "time"

type OrderRequest struct {
	Items []OrderItemRequest `json:"items" binding:"required,min=1"`
	CouponCode string `json:"couponCode,omitempty"`
}



type OrderItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity int `json:"quantity" binding:"required,min=1"`
}



type OrderResponse struct {
	ID string `json:"id"`
	CouponCode string `json:"couponCode,omitempty"`
	Items []OrderItemRequest `json:"items"`
	Products []ProductResponse `json:"products"`
	TotalPrice float64 `json:"totalPrice"`
	Discount float64 `json:"discount"`
	FinalPrice float64 `json:"finalPrice"`
	CreatedAt time.Time `json:"createdAt"`
}
