// Package store provides storage implementations for the inventory system.
package store

import (
	"aexp_assesment/domain"
	"context"
	"fmt"
	"sort"
	"sync"
)

// InMemoryStore is a thread-safe in-memory for domain.ProductStore
type InMemoryStore struct {
	mu       sync.RWMutex
	products map[string]domain.Product
}

// NewInMemoryStore constructs a new InMemoryStore
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		products: make(map[string]domain.Product),
	}
}

// compile-time assertion that InMemoryStore implements domain.ProductStore
var _ domain.ProductStore = (*InMemoryStore)(nil)

func (s *InMemoryStore) Create(ctx context.Context, product domain.Product) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	//validations for empty product ID, name, negative price or quantity
	if product.ID == "" {
		return domain.NewInvalidProductError("id", "cannot be empty", product.ID)
	}
	if product.Name == "" {
		return domain.NewInvalidProductError("name", "cannot be empty", product.Name)
	}
	if product.Price < 0 {
		return domain.NewInvalidProductError("price", "must be non-negative", product.Price)
	}
	if product.Quantity < 0 {
		return domain.NewInvalidProductError("quantity", "must be non-negative", product.Quantity)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[product.ID]; exists {
		return domain.NewDuplicateProductError(product.ID)
	}
	s.products[product.ID] = product
	return nil
}

func (s *InMemoryStore) Get(ctx context.Context, id string) (domain.Product, error) {
	select {
	case <-ctx.Done():
		return domain.Product{}, ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.products[id]
	if !ok {
		return domain.Product{}, domain.NewProductNotFoundError(id)
	}
	return p, nil
}

func (s *InMemoryStore) Update(ctx context.Context, id string, product domain.Product) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if product.Name == "" {
		return domain.NewInvalidProductError("name", "cannot be empty", product.Name)
	}
	if product.Price < 0 {
		return domain.NewInvalidProductError("price", "must be non-negative", product.Price)
	}
	if product.Quantity < 0 {
		return domain.NewInvalidProductError("quantity", "must be non-negative", product.Quantity)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.products[id]; !ok {
		return domain.NewProductNotFoundError(id)
	}
	product.ID = id
	s.products[id] = product
	return nil
}

func (s *InMemoryStore) Delete(ctx context.Context, id string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.products[id]; !ok {
		return domain.NewProductNotFoundError(id)
	}
	delete(s.products, id)
	return nil
}

func (s *InMemoryStore) List(ctx context.Context, filter domain.ListFilter) ([]domain.Product, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]domain.Product, 0, len(s.products))
	for _, p := range s.products {
		if filter.Category != "" && p.Category != filter.Category {
			continue
		}
		if filter.MinPrice != nil && p.Price < *filter.MinPrice {
			continue
		}
		if filter.MaxPrice != nil && p.Price > *filter.MaxPrice {
			continue
		}
		out = append(out, p)
	}

	switch filter.SortBy {
	case "name":
		sort.Slice(out, func(i, j int) bool {
			if filter.Order == "desc" {
				return out[i].Name > out[j].Name
			}
			return out[i].Name < out[j].Name
		})
	case "price":
		sort.Slice(out, func(i, j int) bool {
			if filter.Order == "desc" {
				return out[i].Price > out[j].Price
			}
			return out[i].Price < out[j].Price
		})
	case "quantity":
		sort.Slice(out, func(i, j int) bool {
			if filter.Order == "desc" {
				return out[i].Quantity > out[j].Quantity
			}
			return out[i].Quantity < out[j].Quantity
		})
	}

	return out, nil
}

func (s *InMemoryStore) BulkImport(ctx context.Context, products []domain.Product) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	const maxWorkers = 10
	if len(products) == 0 {
		return nil
	}

	type result struct {
		id  string
		err error
	}

	jobs := make(chan domain.Product)
	results := make(chan result, len(products))

	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case p, ok := <-jobs:
				if !ok {
					return
				}
				if err := s.Create(ctx, p); err != nil {
					results <- result{id: p.ID, err: fmt.Errorf("id=%s: %w", p.ID, err)}
				} else {
					results <- result{id: p.ID, err: nil}
				}
			}
		}
	}

	nWorkers := maxWorkers
	if len(products) < nWorkers {
		nWorkers = len(products)
	}

	wg.Add(nWorkers)
	for i := 0; i < nWorkers; i++ {
		go worker()
	}

	// feed jobs
	go func() {
		defer close(jobs)
		for _, p := range products {
			select {
			case <-ctx.Done():
				return
			case jobs <- p:
			}
		}
	}()

	// collect results
	var collected error
	received := 0
	for received < len(products) {
		select {
		case <-ctx.Done():
			// wait for workers to stop then return context error
			wg.Wait()
			return ctx.Err()
		case res := <-results:
			received++
			if res.err != nil {
				if collected == nil {
					collected = res.err
				} else {
					collected = fmt.Errorf("%v; %w", collected, res.err)
				}
			}
		}
	}

	// all results received; wait for workers
	wg.Wait()
	return collected
}
