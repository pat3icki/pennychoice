// errors/errors.go
package errors

import (
	"errors"
	"fmt"
)

// ErrorType represents the category of error
type ErrorType string

const (
	TypeDatabase   ErrorType = "database"
	TypeValidation ErrorType = "validation"
	TypeNotFound   ErrorType = "not_found"
	TypePermission ErrorType = "permission"
	TypeConflict   ErrorType = "conflict"
	TypeInternal   ErrorType = "internal"
	TypeInput      ErrorType = "input"
)

// AppError is the base error interface
type AppError interface {
	error
	Type() ErrorType
	Package() string
	Function() string
	Unwrap() error
}

// DatabaseError represents database-related errors
type DatabaseError struct {
	pkg string
	fn  string
	err error
}

func (e DatabaseError) Error() string {
	return fmt.Sprintf("database error [%s.%s]: %v", e.pkg, e.fn, e.err)
}

func (e DatabaseError) Type() ErrorType  { return TypeDatabase }
func (e DatabaseError) Package() string  { return e.pkg }
func (e DatabaseError) Function() string { return e.fn }
func (e DatabaseError) Unwrap() error    { return e.err }

// NewDatabaseError creates a new database error
func NewDatabaseError(pkg, fn string, err error) error {
	return DatabaseError{
		pkg: pkg,
		fn:  fn,
		err: err,
	}
}

// ValidationError represents validation errors
type ValidationError struct {
	pkg     string
	fn      string
	Field   string
	Message string
	err     error
}

func (e ValidationError) Error() string {
	msg := fmt.Sprintf("validation error [%s.%s]", e.pkg, e.fn)
	if e.Field != "" {
		msg += fmt.Sprintf(" field '%s'", e.Field)
	}
	if e.Message != "" {
		msg += fmt.Sprintf(": %s", e.Message)
	}
	if e.err != nil {
		msg += fmt.Sprintf(": %v", e.err)
	}
	return msg
}

func (e ValidationError) Type() ErrorType  { return TypeValidation }
func (e ValidationError) Package() string  { return e.pkg }
func (e ValidationError) Function() string { return e.fn }
func (e ValidationError) Unwrap() error    { return e.err }

// NewValidationError creates a new validation error
func NewValidationError(pkg, fn, field, message string) error {
	return ValidationError{
		pkg:     pkg,
		fn:      fn,
		Field:   field,
		Message: message,
	}
}

// NewValidationErrorWithCause creates a validation error with a cause
func NewValidationErrorWithCause(pkg, fn, field, message string, err error) error {
	return ValidationError{
		pkg:     pkg,
		fn:      fn,
		Field:   field,
		Message: message,
		err:     err,
	}
}

// NotFoundError represents "not found" errors
type NotFoundError struct {
	pkg      string
	fn       string
	Resource string
	ID       string
	err      error
}

func (e NotFoundError) Error() string {
	msg := fmt.Sprintf("not found error [%s.%s]", e.pkg, e.fn)
	if e.Resource != "" {
		msg += fmt.Sprintf(" resource '%s'", e.Resource)
	}
	if e.ID != "" {
		msg += fmt.Sprintf(" id '%s'", e.ID)
	}
	if e.err != nil {
		msg += fmt.Sprintf(": %v", e.err)
	}
	return msg
}

func (e NotFoundError) Type() ErrorType  { return TypeNotFound }
func (e NotFoundError) Package() string  { return e.pkg }
func (e NotFoundError) Function() string { return e.fn }
func (e NotFoundError) Unwrap() error    { return e.err }

// NewNotFoundError creates a new not found error
func NewNotFoundError(pkg, fn, resource, id string) error {
	return NotFoundError{
		pkg:      pkg,
		fn:       fn,
		Resource: resource,
		ID:       id,
	}
}

// PermissionError represents permission/authorization errors
type PermissionError struct {
	pkg    string
	fn     string
	UserID string
	Action string
	err    error
}

func (e PermissionError) Error() string {
	msg := fmt.Sprintf("permission error [%s.%s]", e.pkg, e.fn)
	if e.UserID != "" {
		msg += fmt.Sprintf(" user '%s'", e.UserID)
	}
	if e.Action != "" {
		msg += fmt.Sprintf(" action '%s'", e.Action)
	}
	if e.err != nil {
		msg += fmt.Sprintf(": %v", e.err)
	}
	return msg
}

func (e PermissionError) Type() ErrorType  { return TypePermission }
func (e PermissionError) Package() string  { return e.pkg }
func (e PermissionError) Function() string { return e.fn }
func (e PermissionError) Unwrap() error    { return e.err }

// NewPermissionError creates a new permission error
func NewPermissionError(pkg, fn, userID, action string) error {
	return PermissionError{
		pkg:    pkg,
		fn:     fn,
		UserID: userID,
		Action: action,
	}
}

// ConflictError represents conflicts (e.g., duplicate entries)
type ConflictError struct {
	pkg      string
	fn       string
	Resource string
	Message  string
	err      error
}

func (e ConflictError) Error() string {
	msg := fmt.Sprintf("conflict error [%s.%s]", e.pkg, e.fn)
	if e.Resource != "" {
		msg += fmt.Sprintf(" resource '%s'", e.Resource)
	}
	if e.Message != "" {
		msg += fmt.Sprintf(": %s", e.Message)
	}
	if e.err != nil {
		msg += fmt.Sprintf(": %v", e.err)
	}
	return msg
}

func (e ConflictError) Type() ErrorType  { return TypeConflict }
func (e ConflictError) Package() string  { return e.pkg }
func (e ConflictError) Function() string { return e.fn }
func (e ConflictError) Unwrap() error    { return e.err }

// NewConflictError creates a new conflict error
func NewConflictError(pkg, fn, resource, message string) error {
	return ConflictError{
		pkg:      pkg,
		fn:       fn,
		Resource: resource,
		Message:  message,
	}
}

// InputError represents input/argument errors
type InputError struct {
	pkg      string
	fn       string
	Argument string
	Message  string
	err      error
}

func (e InputError) Error() string {
	msg := fmt.Sprintf("input error [%s.%s]", e.pkg, e.fn)
	if e.Argument != "" {
		msg += fmt.Sprintf(" argument '%s'", e.Argument)
	}
	if e.Message != "" {
		msg += fmt.Sprintf(": %s", e.Message)
	}
	if e.err != nil {
		msg += fmt.Sprintf(": %v", e.err)
	}
	return msg
}

func (e InputError) Type() ErrorType  { return TypeInput }
func (e InputError) Package() string  { return e.pkg }
func (e InputError) Function() string { return e.fn }
func (e InputError) Unwrap() error    { return e.err }

// NewInputError creates a new input error
func NewInputError(pkg, fn, argument, message string) error {
	return InputError{
		pkg:      pkg,
		fn:       fn,
		Argument: argument,
		Message:  message,
	}
}

// InternalError represents unexpected internal errors
type InternalError struct {
	pkg     string
	fn      string
	Message string
	err     error
}

func (e InternalError) Error() string {
	msg := fmt.Sprintf("internal error [%s.%s]", e.pkg, e.fn)
	if e.Message != "" {
		msg += fmt.Sprintf(": %s", e.Message)
	}
	if e.err != nil {
		msg += fmt.Sprintf(": %v", e.err)
	}
	return msg
}

func (e InternalError) Type() ErrorType  { return TypeInternal }
func (e InternalError) Package() string  { return e.pkg }
func (e InternalError) Function() string { return e.fn }
func (e InternalError) Unwrap() error    { return e.err }

// NewInternalError creates a new internal error
func NewInternalError(pkg, fn, message string, err error) error {
	return InternalError{
		pkg:     pkg,
		fn:      fn,
		Message: message,
		err:     err,
	}
}

// Helper functions to check error types using errors.As (idiomatic)
func IsDatabaseError(err error) bool {
	var target DatabaseError
	return errors.As(err, &target)
}

func IsNotFoundError(err error) bool {
	var target NotFoundError
	return errors.As(err, &target)
}

func IsValidationError(err error) bool {
	var target ValidationError
	return errors.As(err, &target)
}

func IsPermissionError(err error) bool {
	var target PermissionError
	return errors.As(err, &target)
}

func IsConflictError(err error) bool {
	var target ConflictError
	return errors.As(err, &target)
}

func IsInputError(err error) bool {
	var target InputError
	return errors.As(err, &target)
}

func IsInternalError(err error) bool {
	var target InternalError
	return errors.As(err, &target)
}

// GetType extracts the error type from an AppError
func GetType(err error) ErrorType {
	if appErr, ok := err.(AppError); ok {
		return appErr.Type()
	}
	return TypeInternal
}

// Wrap wraps an error with additional context
func Wrap(err error, pkg, fn string) error {
	if err == nil {
		return nil
	}

	if appErr, ok := err.(AppError); ok {
		return appErr
	}

	return InternalError{
		pkg: pkg,
		fn:  fn,
		err: err,
	}
}
