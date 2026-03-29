CREATE TABLE IF NOT EXISTS orders (
	id TEXT PRIMARY KEY,
	coupon_code TEXT NOT NULL DEFAULT '',
	total_price DOUBLE PRECISION NOT NULL,
	discount DOUBLE PRECISION NOT NULL DEFAULT 0,
	final_price DOUBLE PRECISION NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
	order_id TEXT NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
	product_id TEXT NOT NULL REFERENCES products (id),
	quantity INTEGER NOT NULL CHECK (quantity > 0),
	PRIMARY KEY (order_id, product_id)
);

CREATE INDEX IF NOT EXISTS order_items_product_id_idx ON order_items (product_id);
