package orm

import (
	"fmt"
)

type DoesNotExist struct {
	ModelName string
	Filter   string
}

func (e *DoesNotExist) Error() string {
	return fmt.Sprintf("orm: %s matching query does not exist: %s", e.ModelName, e.Filter)
}

type MultipleObjectsReturned struct {
	ModelName string
	Filter   string
}

func (e *MultipleObjectsReturned) Error() string {
	return fmt.Sprintf("orm: %s matching query returned multiple objects: %s", e.ModelName, e.Filter)
}

type FieldError struct {
	ModelName string
	FieldName string
	Message   string
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("orm: %s.%s: %s", e.ModelName, e.FieldName, e.Message)
}

type ValidationError struct {
	Field   string
	Message string
	Code    string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("orm: validation error on %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("orm: validation error: %s", e.Message)
}

type IntegrityError struct {
	Constraint string
	Message    string
}

func (e *IntegrityError) Error() string {
	return fmt.Sprintf("orm: integrity error: %s: %s", e.Constraint, e.Message)
}

type TransactionError struct {
	Op      string
	Message string
}

func (e *TransactionError) Error() string {
	return fmt.Sprintf("orm: transaction error (%s): %s", e.Op, e.Message)
}