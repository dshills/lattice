package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaxProjectNameLen        = 200
	MaxProjectDescriptionLen = 5000

	// DefaultProjectID is the well-known ID for the auto-created default project.
	DefaultProjectID = "00000000-0000-0000-0000-000000000001"
)

// Project groups related work items into a single namespace.
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Validate checks field constraints for a Project.
func (p *Project) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if utf8.RuneCountInString(p.Name) > MaxProjectNameLen {
		return fmt.Errorf("%w: name exceeds %d characters", ErrInvalidInput, MaxProjectNameLen)
	}
	if utf8.RuneCountInString(p.Description) > MaxProjectDescriptionLen {
		return fmt.Errorf("%w: description exceeds %d characters", ErrInvalidInput, MaxProjectDescriptionLen)
	}
	return nil
}
