package domain

import (
	"errors"
	"testing"
)

func TestProductNotFoundError(t *testing.T) {
	t.Run("Error message formatting", func(t *testing.T) {
		err := NewProductNotFoundError("prod-123")
		expected := "product not found: id=prod-123"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("errors.Is detection", func(t *testing.T) {
		err := NewProductNotFoundError("prod-123")
		target := &ProductNotFoundError{}
		if !errors.Is(err, target) {
			t.Error("errors.Is should detect ProductNotFoundError")
		}
	})

	t.Run("errors.As conversion", func(t *testing.T) {
		err := NewProductNotFoundError("prod-456")
		var pnf *ProductNotFoundError
		if !errors.As(err, &pnf) {
			t.Fatal("errors.As should convert to ProductNotFoundError")
		}
		if pnf.ProductID != "prod-456" {
			t.Errorf("expected ProductID prod-456, got %s", pnf.ProductID)
		}
	})

	t.Run("IsProductNotFoundError helper", func(t *testing.T) {
		err := NewProductNotFoundError("prod-789")
		if !IsProductNotFoundError(err) {
			t.Error("IsProductNotFoundError should return true")
		}
	})
}

func TestInvalidProductError(t *testing.T) {
	t.Run("Error message formatting", func(t *testing.T) {
		err := NewInvalidProductError("price", "must be positive", -10.5)
		expected := "invalid product: field=price, reason=must be positive, value=-10.5"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("errors.Is detection", func(t *testing.T) {
		err := NewInvalidProductError("name", "cannot be empty", "")
		target := &InvalidProductError{}
		if !errors.Is(err, target) {
			t.Error("errors.Is should detect InvalidProductError")
		}
	})

	t.Run("errors.As conversion", func(t *testing.T) {
		err := NewInvalidProductError("quantity", "must be non-negative", -5)
		var ipe *InvalidProductError
		if !errors.As(err, &ipe) {
			t.Fatal("errors.As should convert to InvalidProductError")
		}
		if ipe.Field != "quantity" || ipe.Reason != "must be non-negative" {
			t.Errorf("error fields not correctly preserved")
		}
	})

	t.Run("IsInvalidProductError helper", func(t *testing.T) {
		err := NewInvalidProductError("category", "invalid category", "Unknown")
		if !IsInvalidProductError(err) {
			t.Error("IsInvalidProductError should return true")
		}
	})
}

func TestDuplicateProductError(t *testing.T) {
	t.Run("Error message formatting", func(t *testing.T) {
		err := NewDuplicateProductError("prod-001")
		expected := "duplicate product: id=prod-001 already exists"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("errors.Is detection", func(t *testing.T) {
		err := NewDuplicateProductError("prod-002")
		target := &DuplicateProductError{}
		if !errors.Is(err, target) {
			t.Error("errors.Is should detect DuplicateProductError")
		}
	})

	t.Run("errors.As conversion", func(t *testing.T) {
		err := NewDuplicateProductError("prod-003")
		var dpe *DuplicateProductError
		if !errors.As(err, &dpe) {
			t.Fatal("errors.As should convert to DuplicateProductError")
		}
		if dpe.ProductID != "prod-003" {
			t.Errorf("expected ProductID prod-003, got %s", dpe.ProductID)
		}
	})

	t.Run("IsDuplicateProductError helper", func(t *testing.T) {
		err := NewDuplicateProductError("prod-004")
		if !IsDuplicateProductError(err) {
			t.Error("IsDuplicateProductError should return true")
		}
	})
}

func TestErrorTypeDiscrimination(t *testing.T) {
	t.Run("Different error types are not confused", func(t *testing.T) {
		pnfErr := NewProductNotFoundError("prod-1")
		ipeErr := NewInvalidProductError("price", "negative", -5)
		dpeErr := NewDuplicateProductError("prod-2")

		// ProductNotFoundError checks
		if !IsProductNotFoundError(pnfErr) {
			t.Error("should identify ProductNotFoundError")
		}
		if IsInvalidProductError(pnfErr) {
			t.Error("ProductNotFoundError should not be InvalidProductError")
		}
		if IsDuplicateProductError(pnfErr) {
			t.Error("ProductNotFoundError should not be DuplicateProductError")
		}

		// InvalidProductError checks
		if !IsInvalidProductError(ipeErr) {
			t.Error("should identify InvalidProductError")
		}
		if IsProductNotFoundError(ipeErr) {
			t.Error("InvalidProductError should not be ProductNotFoundError")
		}
		if IsDuplicateProductError(ipeErr) {
			t.Error("InvalidProductError should not be DuplicateProductError")
		}

		// DuplicateProductError checks
		if !IsDuplicateProductError(dpeErr) {
			t.Error("should identify DuplicateProductError")
		}
		if IsProductNotFoundError(dpeErr) {
			t.Error("DuplicateProductError should not be ProductNotFoundError")
		}
		if IsInvalidProductError(dpeErr) {
			t.Error("DuplicateProductError should not be InvalidProductError")
		}
	})
}
