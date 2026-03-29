package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderLine struct {
	ProductID string
	Quantity  int
}

type OrderCreateInput struct {
	ID          string
	CouponCode  string
	TotalPrice  float64
	Discount    float64
	FinalPrice  float64
	Lines       []OrderLine
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, in OrderCreateInput) (time.Time, error)
}

type orderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) CreateOrder(ctx context.Context, in OrderCreateInput) (time.Time, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return time.Time{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var createdAt time.Time
	err = tx.QueryRow(ctx, `
INSERT INTO orders (id, coupon_code, total_price, discount, final_price)
VALUES ($1, $2, $3, $4, $5)
RETURNING created_at
`, in.ID, in.CouponCode, in.TotalPrice, in.Discount, in.FinalPrice).Scan(&createdAt)
	if err != nil {
		return time.Time{}, fmt.Errorf("insert order: %w", err)
	}

	for _, line := range in.Lines {
		if _, err := tx.Exec(ctx, `
INSERT INTO order_items (order_id, product_id, quantity)
VALUES ($1, $2, $3)
`, in.ID, line.ProductID, line.Quantity); err != nil {
			return time.Time{}, fmt.Errorf("insert order_item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return time.Time{}, fmt.Errorf("commit: %w", err)
	}
	return createdAt, nil
}
