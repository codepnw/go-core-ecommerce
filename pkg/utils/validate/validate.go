package validate

import "github.com/go-playground/validator/v10"

var v = validator.New()

func Struct(input any) error {
	return v.Struct(input)
}
