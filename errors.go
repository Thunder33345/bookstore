package bookstore

import (
	"errors"
	"fmt"
)

var MissingIDError = errors.New("missing id")

type NoResultError struct {
	resource string
}

func NewNoResultError(resource string) error {
	return &NoResultError{
		resource: resource,
	}
}

func (e *NoResultError) Error() string {
	return fmt.Sprintf("%s does not exist", e.resource)
}

type DuplicateError struct {
	resourceType string
}

func NewDuplicateError(resourceType string) error {
	return &DuplicateError{
		resourceType: resourceType,
	}
}

func (e *DuplicateError) Error() string {
	return fmt.Sprintf(`%s already exist`, e.resourceType)
}

type DependedError struct {
	resource string
}

func NewDependedError(resource string) error {
	return &DependedError{
		resource: resource,
	}
}
func (e *DependedError) Error() string {
	return fmt.Sprintf("%s is being depended by other books", e.resource)
}

type InvalidDependencyError struct {
	resource string
}

func NewInvalidDependencyError(resource string) error {
	return &InvalidDependencyError{
		resource: resource,
	}
}

func (e *InvalidDependencyError) Error() string {
	return fmt.Sprintf("invalid value on books.%s", e.resource)
}
