package dto


type ProductResponse struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Price float64 `json:"price"`
	Category string `json:"category"`
	Image string `json:"image"`
}