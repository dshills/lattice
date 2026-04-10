package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrInvalidTransition = errors.New("invalid state transition")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidInput      = errors.New("invalid input")
	ErrValidation        = errors.New("validation error")
	ErrConflict          = errors.New("conflict")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrDuplicateEmail    = fmt.Errorf("%w: duplicate email", ErrConflict)
)
