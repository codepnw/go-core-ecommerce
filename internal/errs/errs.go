package errs

import "errors"

// Error Users
var (
	ErrUserNotFound           = errors.New("user not found")
	ErrEmailAlreadyExists     = errors.New("email already exists")
	ErrInvalidEmailOrPassword = errors.New("invalid email or password")
	ErrTokenNotFound          = errors.New("token not found")
	ErrTokenRevoked           = errors.New("token revoked")
	ErrTokenExpires           = errors.New("token expires")
	ErrUnauthorized           = errors.New("unauthorized")
)

// Error Products
var (
	ErrProductNotFound  = errors.New("product not found")
	ErrProductSKUExists = errors.New("product sku already exists")
	ErrStockNotEnough   = errors.New("stock not enough")
)
