// Package domain defines core business types and interfaces.
package domain

import "context"

// Product represents an inventory product
type Product struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Category string  `json:"category"`
}

// ListFilter allows filtering and sorting results from List
type ListFilter struct {
	Category string
	MinPrice *float64
	MaxPrice *float64
	SortBy   string // "name", "price", "quantity"
	Order    string // "asc" or "desc"
}

// ProductStore defines the storage interface for products
type ProductStore interface {
	Create(ctx context.Context, product Product) error
	Get(ctx context.Context, id string) (Product, error)
	Update(ctx context.Context, id string, product Product) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter ListFilter) ([]Product, error)
	BulkImport(ctx context.Context, products []Product) error
}
