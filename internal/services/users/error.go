package users

import "errors"

// ErrInvalidCredentials prevents user enumeration by returning the same error
// whether the user doesn't exist or the password is wrong.
var (
	ErrInvalidCredentials     = errors.New("invalid identifier or password")
	ErrUserAlreadyDeactivated = errors.New("user is already deactivated")
	ErrUserNotFound           = errors.New("user not found")
)

const (
	PackageName string = "service/users"
)

