package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_Validate_Valid(t *testing.T) {
	u := &User{Email: "alice@example.com", DisplayName: "Alice"}
	assert.NoError(t, u.Validate())
}

func TestUser_Validate_EmptyEmail(t *testing.T) {
	u := &User{Email: "", DisplayName: "Alice"}
	assert.ErrorIs(t, u.Validate(), ErrInvalidInput)
}

func TestUser_Validate_InvalidEmail(t *testing.T) {
	u := &User{Email: "not-an-email", DisplayName: "Alice"}
	assert.ErrorIs(t, u.Validate(), ErrInvalidInput)
}

func TestUser_Validate_EmailTooLong(t *testing.T) {
	u := &User{Email: strings.Repeat("a", 310) + "@example.com", DisplayName: "Alice"}
	assert.ErrorIs(t, u.Validate(), ErrInvalidInput)
}

func TestUser_Validate_EmptyDisplayName(t *testing.T) {
	u := &User{Email: "alice@example.com", DisplayName: ""}
	assert.ErrorIs(t, u.Validate(), ErrInvalidInput)
}

func TestUser_Validate_DisplayNameTooLong(t *testing.T) {
	u := &User{Email: "alice@example.com", DisplayName: strings.Repeat("a", MaxDisplayNameLen+1)}
	assert.ErrorIs(t, u.Validate(), ErrInvalidInput)
}

func TestUser_Validate_MaxLengthDisplayName(t *testing.T) {
	u := &User{Email: "alice@example.com", DisplayName: strings.Repeat("a", MaxDisplayNameLen)}
	assert.NoError(t, u.Validate())
}

func TestValidatePassword_Valid(t *testing.T) {
	assert.NoError(t, ValidatePassword("password123"))
}

func TestValidatePassword_TooShort(t *testing.T) {
	assert.ErrorIs(t, ValidatePassword("short"), ErrInvalidInput)
}

func TestValidatePassword_ExactMinLength(t *testing.T) {
	assert.NoError(t, ValidatePassword(strings.Repeat("a", MinPasswordLen)))
}

func TestValidatePassword_OneBelowMin(t *testing.T) {
	assert.ErrorIs(t, ValidatePassword(strings.Repeat("a", MinPasswordLen-1)), ErrInvalidInput)
}
