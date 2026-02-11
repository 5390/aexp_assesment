package store

import (
	"aexp_assesment/domain"
	"context"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestCreateValidation_TableDriven(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()

	cases := []struct {
		name    string
		product domain.Product
		wantErr bool
	}{
		{"empty id", domain.Product{ID: "", Name: "A", Price: 1, Quantity: 1}, true},
		{"empty name", domain.Product{ID: "x1", Name: "", Price: 1, Quantity: 1}, true},
		{"negative price", domain.Product{ID: "x2", Name: "A", Price: -1, Quantity: 1}, true},
		{"negative quantity", domain.Product{ID: "x3", Name: "A", Price: 1, Quantity: -5}, true},
		{"valid", domain.Product{ID: "x4", Name: "A", Price: 1, Quantity: 0}, false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := s.Create(ctx, tc.product)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for case %s", tc.name)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetUpdateDelete_NotFoundAndInvalid(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()

	t.Run("get not found", func(t *testing.T) {
		_, err := s.Get(ctx, "no-such")
		if !domain.IsProductNotFoundError(err) {
			t.Fatalf("expected ProductNotFoundError, got %v", err)
		}
	})

	t.Run("update not found", func(t *testing.T) {
		err := s.Update(ctx, "no-such", domain.Product{Name: "A", Price: 1, Quantity: 1})
		if !domain.IsProductNotFoundError(err) {
			t.Fatalf("expected ProductNotFoundError, got %v", err)
		}
	})

	t.Run("delete not found", func(t *testing.T) {
		err := s.Delete(ctx, "no-such")
		if !domain.IsProductNotFoundError(err) {
			t.Fatalf("expected ProductNotFoundError, got %v", err)
		}
	})

	// create and attempt invalid update
	if err := s.Create(ctx, domain.Product{ID: "u1", Name: "V", Price: 2, Quantity: 1}); err != nil {
		t.Fatalf("setup create failed: %v", err)
	}
	t.Run("update invalid", func(t *testing.T) {
		if err := s.Update(ctx, "u1", domain.Product{Name: "", Price: 1, Quantity: 1}); !domain.IsInvalidProductError(err) {
			t.Fatalf("expected InvalidProductError, got %v", err)
		}
	})
}

func TestListSortingAndFiltering(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()
	_ = s.Create(ctx, domain.Product{ID: "a", Name: "Alpha", Price: 5, Quantity: 3, Category: "C1"})
	_ = s.Create(ctx, domain.Product{ID: "b", Name: "Beta", Price: 2, Quantity: 7, Category: "C2"})
	_ = s.Create(ctx, domain.Product{ID: "c", Name: "Gamma", Price: 9, Quantity: 1, Category: "C1"})

	t.Run("filter by category", func(t *testing.T) {
		out, err := s.List(ctx, domain.ListFilter{Category: "C1"})
		if err != nil {
			t.Fatalf("list failed: %v", err)
		}
		if len(out) != 2 {
			t.Fatalf("expected 2, got %d", len(out))
		}
	})

	t.Run("sort by price desc", func(t *testing.T) {
		out, _ := s.List(ctx, domain.ListFilter{SortBy: "price", Order: "desc"})
		if len(out) < 3 || out[0].Price < out[1].Price {
			t.Fatalf("unexpected sort order by price desc")
		}
	})
}

func TestBulkImport_ErrorsAndCancellation(t *testing.T) {
	s := NewInMemoryStore()

	// duplicate IDs should produce error collection
	products := []domain.Product{
		{ID: "d1", Name: "A", Price: 1, Quantity: 1},
		{ID: "d1", Name: "A", Price: 1, Quantity: 1},
	}
	ctx := context.Background()
	err := s.BulkImport(ctx, products)
	if err == nil {
		t.Fatalf("expected error due to duplicate IDs")
	}

	// cancellation propagated
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.BulkImport(canceledCtx, []domain.Product{{ID: "x1", Name: "N", Price: 1, Quantity: 1}}); err == nil {
		t.Fatalf("expected context error on canceled context")
	}
}

func TestInMemoryStore_ConcurrentAccess(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()
	var wg sync.WaitGroup

	// create many products concurrently
	n := 100
	wg.Add(n)
	for i := 0; i < n; i++ {
		id := "p-conc-" + strconv.Itoa(i)
		go func(id string) {
			defer wg.Done()
			_ = s.Create(ctx, domain.Product{ID: id, Name: "X", Price: 1.0, Quantity: 1, Category: "C"})
			_, _ = s.Get(ctx, id)
		}(id)
	}
	wg.Wait()

	out, err := s.List(ctx, domain.ListFilter{})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(out) != n {
		t.Fatalf("expected %d products, got %d", n, len(out))
	}
}

func TestBulkImport_Timeout(t *testing.T) {
	s := NewInMemoryStore()
	// large number to ensure work takes some time
	n := 1000
	products := make([]domain.Product, 0, n)
	for i := 0; i < n; i++ {
		products = append(products, domain.Product{ID: "t-" + strconv.Itoa(i), Name: "X", Price: 1.0, Quantity: 1, Category: "C"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	err := s.BulkImport(ctx, products)
	if err == nil {
		t.Fatalf("expected timeout or cancellation error, got nil")
	}
}

func BenchmarkInMemoryStore_Create(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := NewInMemoryStore()
		p := domain.Product{ID: "b-create-" + strconv.Itoa(i), Name: "Bench", Price: 1, Quantity: 1}
		_ = s.Create(context.Background(), p)
	}
}

func BenchmarkInMemoryStore_Get(b *testing.B) {
	s := NewInMemoryStore()
	for i := 0; i < 1000; i++ {
		_ = s.Create(context.Background(), domain.Product{ID: "b-get-" + strconv.Itoa(i), Name: "X", Price: 1, Quantity: 1})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := "b-get-" + strconv.Itoa(i%1000)
		_, _ = s.Get(context.Background(), id)
	}
}
