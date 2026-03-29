package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type seedProduct struct {
	id, name, category, image string
	price                       float64
}

// Unsplash CDN URLs — food photos, sized for apps (format & width).
var seedProducts = []seedProduct{
	{"seed-001", "Margherita Pizza", "Italian", "https://images.unsplash.com/photo-1565299624946-b28f40a0ae38?auto=format&fit=crop&w=800&q=80", 12.99},
	{"seed-002", "Classic Cheeseburger", "American", "https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&w=800&q=80", 10.50},
	{"seed-003", "Salmon Nigiri Set", "Japanese", "https://images.unsplash.com/photo-1579584427862-fadc1a2d9d01?auto=format&fit=crop&w=800&q=80", 18.00},
	{"seed-004", "Creamy Penne Alfredo", "Italian", "https://images.unsplash.com/photo-1621996346565-e3dbc646d9a9?auto=format&fit=crop&w=800&q=80", 14.25},
	{"seed-005", "Street Tacos (3)", "Mexican", "https://images.unsplash.com/photo-1565299585323-381d8b46c791?auto=format&fit=crop&w=800&q=80", 9.75},
	{"seed-006", "Tonkotsu Ramen", "Japanese", "https://images.unsplash.com/photo-1569718212169-50389bbb506e?auto=format&fit=crop&w=800&q=80", 15.50},
	{"seed-007", "Grilled Caesar Salad", "Salads", "https://images.unsplash.com/photo-1546793665-c74683a3398e?auto=format&fit=crop&w=800&q=80", 8.99},
	{"seed-008", "Crispy Fried Chicken", "American", "https://images.unsplash.com/photo-1626082927389-6cd097cdc6ec?auto=format&fit=crop&w=800&q=80", 13.49},
	{"seed-009", "Ribeye Steak", "Grill", "https://images.unsplash.com/photo-1544025162-d76694265947?auto=format&fit=crop&w=800&q=80", 29.99},
	{"seed-010", "Vanilla Bean Ice Cream", "Desserts", "https://images.unsplash.com/photo-1563805042-768885cd1200?auto=format&fit=crop&w=800&q=80", 5.50},
	{"seed-011", "Berry Pancake Stack", "Breakfast", "https://images.unsplash.com/photo-1567620905732-2d1ec7b2a93b?auto=format&fit=crop&w=800&q=80", 11.00},
	{"seed-012", "Loaded Burrito", "Mexican", "https://images.unsplash.com/photo-1561651824-9e648e6c8c25?auto=format&fit=crop&w=800&q=80", 11.75},
	{"seed-013", "Pad Thai", "Thai", "https://images.unsplash.com/photo-1559314809-723d3c46e05f?auto=format&fit=crop&w=800&q=80", 13.00},
	{"seed-014", "Fish & Chips", "British", "https://images.unsplash.com/photo-1579208570378-8c970854bc28?auto=format&fit=crop&w=800&q=80", 14.50},
	{"seed-015", "Chocolate Layer Cake", "Desserts", "https://images.unsplash.com/photo-1578985545062-69928b1d9587?auto=format&fit=crop&w=800&q=80", 7.25},
}

func seedProductsIfEmpty(ctx context.Context, pool *pgxpool.Pool) error {
	var n int64
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM products`).Scan(&n); err != nil {
		return fmt.Errorf("count products: %w", err)
	}
	if n > 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, p := range seedProducts {
		batch.Queue(
			`INSERT INTO products (id, name, price, category, image) VALUES ($1, $2, $3, $4, $5)`,
			p.id, p.name, p.price, p.category, p.image,
		)
	}
	br := pool.SendBatch(ctx, batch)
	defer br.Close()

	for range seedProducts {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("seed insert: %w", err)
		}
	}
	return nil
}
