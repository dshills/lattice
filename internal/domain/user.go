package domain

import (
	"fmt"
	"net/mail"
	"time"
	"unicode/utf8"
)

const (
	MaxEmailLen       = 320
	MaxDisplayNameLen = 100
	MinPasswordLen    = 8
)

// User represents a registered user in the system.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Validate checks field constraints for a User.
func (u *User) Validate() error {
	if u.Email == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidInput)
	}
	if utf8.RuneCountInString(u.Email) > MaxEmailLen {
		return fmt.Errorf("%w: email exceeds %d characters", ErrInvalidInput, MaxEmailLen)
	}
	if _, err := mail.ParseAddress(u.Email); err != nil {
		return fmt.Errorf("%w: invalid email format", ErrInvalidInput)
	}
	if u.DisplayName == "" {
		return fmt.Errorf("%w: display_name is required", ErrInvalidInput)
	}
	if utf8.RuneCountInString(u.DisplayName) > MaxDisplayNameLen {
		return fmt.Errorf("%w: display_name exceeds %d characters", ErrInvalidInput, MaxDisplayNameLen)
	}
	return nil
}

// ValidatePassword checks that a plaintext password meets minimum requirements.
func ValidatePassword(plain string) error {
	if utf8.RuneCountInString(plain) < MinPasswordLen {
		return fmt.Errorf("%w: password must be at least %d characters", ErrInvalidInput, MinPasswordLen)
	}
	return nil
}
