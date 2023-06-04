package bookstore

import (
	"errors"
	"fmt"
)

var ErrMissingID = errors.New("missing id")

type NoResultError struct {
	resource string
	err      error
}

func NewNoResultError(resource string, err error) error {
	return &NoResultError{
		resource: resource,
		err:      err,
	}
}

func (e *NoResultError) Error() string {
	return fmt.Sprintf("%s does not exist", e.resource)
}

func (e *NoResultError) Unwrap() error {
	return e.err
}

type DuplicateError struct {
	resourceType string
	err          error
}

func NewDuplicateError(resourceType string, err error) error {
	return &DuplicateError{
		resourceType: resourceType,
		err:          err,
	}
}

func (e *DuplicateError) Error() string {
	return fmt.Sprintf(`%s already exist`, e.resourceType)
}

func (e *DuplicateError) Unwrap() error {
	return e.err
}

type DependedError struct {
	resource string
	err      error
}

func NewDependedError(resource string, err error) error {
	return &DependedError{
		resource: resource,
		err:      err,
	}
}

func (e *DependedError) Error() string {
	return fmt.Sprintf("%s is being depended by other books", e.resource)
}

func (e *DependedError) Unwrap() error {
	return e.err
}

type InvalidDependencyError struct {
	resource string
	err      error
}

func NewInvalidDependencyError(resource string, err error) error {
	return &InvalidDependencyError{
		resource: resource,
		err:      err,
	}
}

func (e *InvalidDependencyError) Error() string {
	return fmt.Sprintf("invalid value on books.%s", e.resource)
}
func (e *InvalidDependencyError) Unwrap() error {
	return e.err
}

// NonExistentIDError is returned when paging while using an invalid UUID as the after parameter
type NonExistentIDError struct {
	resourceType string
	err          error
}

func NewNonExistentIDError(resourceType string, err error) error {
	return &NonExistentIDError{
		resourceType: resourceType,
		err:          err,
	}
}

func (e *NonExistentIDError) Error() string {
	return fmt.Sprintf(`nonexistent ID given in "after" while paging on %s`, e.resourceType)
}

func (e *NonExistentIDError) Unwrap() error {
	return e.err
}
