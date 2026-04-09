package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProject_Validate_Valid(t *testing.T) {
	p := &Project{Name: "My Project"}
	assert.NoError(t, p.Validate())
}

func TestProject_Validate_EmptyName(t *testing.T) {
	p := &Project{Name: ""}
	assert.ErrorIs(t, p.Validate(), ErrInvalidInput)
}

func TestProject_Validate_WhitespaceName(t *testing.T) {
	p := &Project{Name: "   "}
	assert.ErrorIs(t, p.Validate(), ErrInvalidInput)
}

func TestProject_Validate_NameTooLong(t *testing.T) {
	p := &Project{Name: strings.Repeat("a", MaxProjectNameLen+1)}
	assert.ErrorIs(t, p.Validate(), ErrInvalidInput)
}

func TestProject_Validate_DescriptionTooLong(t *testing.T) {
	p := &Project{Name: "ok", Description: strings.Repeat("x", MaxProjectDescriptionLen+1)}
	assert.ErrorIs(t, p.Validate(), ErrInvalidInput)
}

func TestProject_Validate_MaxLengthName(t *testing.T) {
	p := &Project{Name: strings.Repeat("a", MaxProjectNameLen)}
	assert.NoError(t, p.Validate())
}
