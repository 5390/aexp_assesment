package domain

import (
	"context"
	"testing"
)

func TestValidateProduct(t *testing.T) {
	tests := []struct {
		name        string
		product     Product
		expectError bool
		errField    string
	}{
		{
			name: "valid product",
			product: Product{
				ID:       "1",
				Name:     "Laptop",
				Price:    1000,
				Quantity: 5,
				Category: "Electronics",
			},
			expectError: false,
		},
		{
			name: "empty name",
			product: Product{
				ID:       "2",
				Name:     "",
				Price:    10,
				Quantity: 1,
			},
			expectError: true,
			errField:    "name",
		},
		{
			name: "negative price",
			product: Product{
				ID:       "3",
				Name:     "Book",
				Price:    -1,
				Quantity: 1,
			},
			expectError: true,
			errField:    "price",
		},
		{
			name: "negative quantity",
			product: Product{
				ID:       "4",
				Name:     "Pen",
				Price:    1,
				Quantity: -5,
			},
			expectError: true,
			errField:    "quantity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProduct(tt.product)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}

				ipe, ok := err.(*InvalidProductError)
				if !ok {
					t.Fatalf("expected InvalidProductError, got %T", err)
				}

				if ipe.Field != tt.errField {
					t.Fatalf(
						"expected error field %q, got %q",
						tt.errField,
						ipe.Field,
					)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestProductStructFields(t *testing.T) {
	p := Product{
		ID:       "id",
		Name:     "name",
		Price:    10.5,
		Quantity: 3,
		Category: "cat",
	}

	if p.ID == "" || p.Name == "" || p.Category == "" {
		t.Fatalf("product fields not set correctly")
	}
}

func TestListFilterZeroValue(t *testing.T) {
	var f ListFilter

	if f.Category != "" {
		t.Fatalf("expected empty category")
	}
	if f.MinPrice != nil {
		t.Fatalf("expected nil MinPrice")
	}
	if f.MaxPrice != nil {
		t.Fatalf("expected nil MaxPrice")
	}
	if f.SortBy != "" || f.Order != "" {
		t.Fatalf("expected empty sort fields")
	}
}

// ---- Interface compile-time test ----

// mockProductStore ensures ProductStore interface stays stable
type mockProductStore struct{}

func (m *mockProductStore) Create(ctx context.Context, p Product) error {
	return nil
}

func (m *mockProductStore) Get(ctx context.Context, id string) (Product, error) {
	return Product{}, nil
}

func (m *mockProductStore) Update(ctx context.Context, id string, p Product) error {
	return nil
}

func (m *mockProductStore) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockProductStore) List(ctx context.Context, f ListFilter) ([]Product, error) {
	return nil, nil
}

func (m *mockProductStore) BulkImport(ctx context.Context, p []Product) error {
	return nil
}

// compile-time assertion
var _ ProductStore = (*mockProductStore)(nil)
