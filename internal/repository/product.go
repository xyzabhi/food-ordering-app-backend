package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xyzabhi/food-ordering-app-backend/internal/model"
)

type ProductRepository interface {
	GetByIDs(ctx context.Context, ids []string) (map[string]model.Product, error)
	ListAll(ctx context.Context) ([]model.Product, error)
}

type productRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) GetByIDs(ctx context.Context, ids []string) (map[string]model.Product, error) {
	if len(ids) == 0 {
		return map[string]model.Product{}, nil
	}
	rows, err := r.db.Query(ctx, `
SELECT id, name, price, category, image, created_at, updated_at
FROM products
WHERE id = ANY($1::text[])
`, ids)
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	out := make(map[string]model.Product)
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Category, &p.Image, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		out[p.ID] = p
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *productRepository) ListAll(ctx context.Context) ([]model.Product, error) {
	rows, err := r.db.Query(ctx, `
SELECT id, name, price, category, image, created_at, updated_at
FROM products
ORDER BY category ASC, name ASC
`)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	var list []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Category, &p.Image, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		list = append(list, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

