package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrInvalidTransition = errors.New("invalid state transition")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidInput      = errors.New("invalid input")
	ErrValidation        = errors.New("validation error")
)
