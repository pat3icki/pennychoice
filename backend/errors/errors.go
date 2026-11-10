package errors

import (
	"fmt"
	"net/http"
)

// ErrorType represents the category of an error.
type ErrorType string

// Error type constants.
const (
	TypeValidation   ErrorType = "validation"
	TypeInput        ErrorType = "input"
	TypeNotFound     ErrorType = "not_found"
	TypePermission   ErrorType = "permission"
	TypeUnauthorized ErrorType = "unauthorized"
	TypeConflict     ErrorType = "conflict"
	TypeDatabase     ErrorType = "database"
	TypeInternal     ErrorType = "internal"
)

// AppError is the interface for all typed application errors.
type AppError interface {
	error
	Type() ErrorType
	Message() string
	Field() string
	Cause() error
	WithCause(error) AppError
	WithField(string) AppError
}

// appError is the concrete implementation.
type appError struct {
	typ   ErrorType
	msg   string
	field string
	cause error
}

func (e *appError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s (cause: %v)", e.typ, e.msg, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.typ, e.msg)
}

func (e *appError) Type() ErrorType { return e.typ }
func (e *appError) Message() string { return e.msg }
func (e *appError) Field() string   { return e.field }
func (e *appError) Cause() error    { return e.cause }
func (e *appError) WithCause(err error) AppError {
	e.cause = err
	return e
}
func (e *appError) WithField(field string) AppError {
	e.field = field
	return e
}

// Constructors for each error type.
func New(typ ErrorType, msg string) AppError {
	return &appError{typ: typ, msg: msg}
}

func Validation(pkg, msg string, err error) AppError {
	return &appError{typ: TypeValidation, msg: fmt.Sprintf("[%s] %s", pkg, msg), cause: err}
}

func Input(pkg, msg string, err error) AppError {
	return &appError{typ: TypeInput, msg: fmt.Sprintf("[%s] %s", pkg, msg), cause: err}
}

func NotFound(pkg, msg string) AppError {
	return &appError{typ: TypeNotFound, msg: fmt.Sprintf("[%s] %s", pkg, msg)}
}

func Permission(pkg, msg string) AppError {
	return &appError{typ: TypePermission, msg: fmt.Sprintf("[%s] %s", pkg, msg)}
}

func Unauthorized(pkg, msg string) AppError {
	return &appError{typ: TypeUnauthorized, msg: fmt.Sprintf("[%s] %s", pkg, msg)}
}

func Conflict(pkg, msg string) AppError {
	return &appError{typ: TypeConflict, msg: fmt.Sprintf("[%s] %s", pkg, msg)}
}

func Database(pkg, msg string, err error) AppError {
	return &appError{typ: TypeDatabase, msg: fmt.Sprintf("[%s] %s", pkg, msg), cause: err}
}

func Internal(pkg, msg string, err error) AppError {
	return &appError{typ: TypeInternal, msg: fmt.Sprintf("[%s] %s", pkg, msg), cause: err}
}

// GetType extracts the ErrorType from an error. Returns TypeInternal if the error is not typed.
func GetType(err error) ErrorType {
	if e, ok := err.(AppError); ok {
		return e.Type()
	}
	return TypeInternal
}

// Is checks if an error is of a given type.
func Is(err error, typ ErrorType) bool {
	return GetType(err) == typ
}

// HTTPStatus maps an error to the appropriate HTTP status code.
func HTTPStatus(err error) int {
	switch GetType(err) {
	case TypeValidation, TypeInput:
		return http.StatusBadRequest
	case TypeNotFound:
		return http.StatusNotFound
	case TypePermission, TypeUnauthorized:
		return http.StatusForbidden
	case TypeConflict:
		return http.StatusConflict
	case TypeDatabase, TypeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// --- Optional: Common sentinel errors (can be extended) ---

var (
	ErrInvalidCredentials = Unauthorized("auth", "invalid credentials")
	ErrUserNotFound       = NotFound("users", "user not found")
	ErrUserAlreadyExists  = Conflict("users", "user already exists")
	ErrCampaignNotFound   = NotFound("campaigns", "campaign not found")
)
