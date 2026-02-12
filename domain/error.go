// Package domain defines error types for the inventory system.
package domain

import (
	"errors"
	"fmt"
)

// ProductNotFoundError is returned when a product with the given ID is not found
type ProductNotFoundError struct {
	ProductID string
}

// Error implements the error interface for ProductNotFoundError
func (e *ProductNotFoundError) Error() string {
	return fmt.Sprintf("product not found: id=%s", e.ProductID)
}

// Is allows proper error type checking with errors.Is()
func (e *ProductNotFoundError) Is(target error) bool {
	_, ok := target.(*ProductNotFoundError)
	return ok
}

// InvalidProductError is returned when product validation fails
type InvalidProductError struct {
	Field  string
	Reason string
	Value  interface{}
}

// Error implements the error interface for InvalidProductError
func (e *InvalidProductError) Error() string {
	return fmt.Sprintf("invalid product: field=%s, reason=%s, value=%v", e.Field, e.Reason, e.Value)
}

// Is allows proper error type checking with errors.Is()
func (e *InvalidProductError) Is(target error) bool {
	_, ok := target.(*InvalidProductError)
	return ok
}

// DuplicateProductError is returned when attempting to create a product with an existing ID
type DuplicateProductError struct {
	ProductID string
}

// Error implements the error interface for DuplicateProductError
func (e *DuplicateProductError) Error() string {
	return fmt.Sprintf("duplicate product: id=%s already exists", e.ProductID)
}

// Is allows proper error type checking with errors.Is()
func (e *DuplicateProductError) Is(target error) bool {
	_, ok := target.(*DuplicateProductError)
	return ok
}

// Helper functions for creating errors with context

// NewProductNotFoundError creates a new ProductNotFoundError
func NewProductNotFoundError(productID string) error {
	return &ProductNotFoundError{ProductID: productID}
}

// NewInvalidProductError creates a new InvalidProductError
func NewInvalidProductError(field, reason string, value interface{}) error {
	return &InvalidProductError{
		Field:  field,
		Reason: reason,
		Value:  value,
	}
}

// NewDuplicateProductError creates a new DuplicateProductError
func NewDuplicateProductError(productID string) error {
	return &DuplicateProductError{ProductID: productID}
}

// Type assertion helpers for use with errors.As()

// IsProductNotFoundError checks if an error is a ProductNotFoundError
func IsProductNotFoundError(err error) bool {
	var pnf *ProductNotFoundError
	return errors.As(err, &pnf)
}

// IsInvalidProductError checks if an error is an InvalidProductError
func IsInvalidProductError(err error) bool {
	var ipe *InvalidProductError
	return errors.As(err, &ipe)
}

// IsDuplicateProductError checks if an error is a DuplicateProductError
func IsDuplicateProductError(err error) bool {
	var dpe *DuplicateProductError
	return errors.As(err, &dpe)
}
