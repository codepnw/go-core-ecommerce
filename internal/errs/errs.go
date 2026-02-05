package errs

import "errors"

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrEmailAlreadyExists     = errors.New("email already exists")
	ErrInvalidEmailOrPassword = errors.New("invalid email or password")
	ErrTokenNotFound          = errors.New("token not found")
	ErrTokenRevoked           = errors.New("token revoked")
	ErrTokenExpires           = errors.New("token expires")
)
