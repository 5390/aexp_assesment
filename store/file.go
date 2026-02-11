package store

import (
	"aexp_assesment/domain"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// FileStore is a JSON file-backed implementation of domain.ProductStore
type FileStore struct {
	mu       sync.RWMutex
	products map[string]domain.Product
	path     string
}

// compile-time assertion
var _ domain.ProductStore = (*FileStore)(nil)

// NewFileStore constructs a FileStore at the given path. If the file exists it will be loaded.
func NewFileStore(path string) (*FileStore, error) {
	s := &FileStore{
		products: make(map[string]domain.Product),
		path:     path,
	}
	if err := s.loadFromFile(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *FileStore) loadFromFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := ioutil.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			// no file yet; that's fine
			return nil
		}
		return err
	}
	var list []domain.Product
	if len(b) == 0 {
		return nil
	}
	if err := json.Unmarshal(b, &list); err != nil {
		return err
	}
	for _, p := range list {
		s.products[p.ID] = p
	}
	return nil
}

func (s *FileStore) saveToFile() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	list := make([]domain.Product, 0, len(s.products))
	for _, p := range s.products {
		list = append(list, p)
	}
	// stable order for deterministic files
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	b, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := ioutil.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *FileStore) Create(ctx context.Context, product domain.Product) error {
	if err := ctx.Err(); err != nil {
		return err
	}
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

	if _, ok := s.products[product.ID]; ok {
		return domain.NewDuplicateProductError(product.ID)
	}
	s.products[product.ID] = product
	return s.saveToFile()
}

func (s *FileStore) Get(ctx context.Context, id string) (domain.Product, error) {
	if err := ctx.Err(); err != nil {
		return domain.Product{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.products[id]
	if !ok {
		return domain.Product{}, domain.NewProductNotFoundError(id)
	}
	return p, nil
}

func (s *FileStore) Update(ctx context.Context, id string, product domain.Product) error {
	if err := ctx.Err(); err != nil {
		return err
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
	return s.saveToFile()
}

func (s *FileStore) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.products[id]; !ok {
		return domain.NewProductNotFoundError(id)
	}
	delete(s.products, id)
	return s.saveToFile()
}

func (s *FileStore) List(ctx context.Context, filter domain.ListFilter) ([]domain.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
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

func (s *FileStore) BulkImport(ctx context.Context, products []domain.Product) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	const maxWorkers = 10
	jobs := make(chan domain.Product)
	errs := make(chan error, len(products))

	var addMu sync.Mutex
	toAdd := make(map[string]domain.Product)

	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for p := range jobs {
			if err := ctx.Err(); err != nil {
				errs <- err
				return
			}
			// validate fields
			if p.ID == "" || p.Name == "" || p.Price < 0 || p.Quantity < 0 {
				errs <- domain.NewInvalidProductError("bulk", "invalid product", p)
				continue
			}
			addMu.Lock()
			if _, exists := toAdd[p.ID]; exists {
				addMu.Unlock()
				errs <- domain.NewDuplicateProductError(p.ID)
				continue
			}
			toAdd[p.ID] = p
			addMu.Unlock()
		}
	}

	nWorkers := maxWorkers
	if len(products) < nWorkers {
		nWorkers = len(products)
	}
	if nWorkers == 0 {
		return nil
	}
	wg.Add(nWorkers)
	for i := 0; i < nWorkers; i++ {
		go worker()
	}

	go func() {
		for _, p := range products {
			select {
			case <-ctx.Done():
				break
			case jobs <- p:
			}
		}
		close(jobs)
	}()

	wg.Wait()
	close(errs)

	var collected error
	for e := range errs {
		if collected == nil {
			collected = e
		} else {
			collected = fmt.Errorf("%v; %w", collected, e)
		}
	}

	// merge toAdd into store with lock, detect duplicates against existing store
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, p := range toAdd {
		if _, exists := s.products[id]; exists {
			e := domain.NewDuplicateProductError(id)
			if collected == nil {
				collected = e
			} else {
				collected = fmt.Errorf("%v; %w", collected, e)
			}
			continue
		}
		s.products[id] = p
	}
	if err := s.saveToFile(); err != nil {
		if collected == nil {
			return err
		}
		return fmt.Errorf("%v; %w", collected, err)
	}
	return collected
}
