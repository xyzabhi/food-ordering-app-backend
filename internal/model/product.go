package model

import "time"


type Product struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Price float64 `json:"price"`
	Category string `json:"category"`
	Image string `json:"image"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}