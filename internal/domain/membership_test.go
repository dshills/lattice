package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidProjectRole_Valid(t *testing.T) {
	assert.True(t, ValidProjectRole(RoleOwner))
	assert.True(t, ValidProjectRole(RoleMember))
	assert.True(t, ValidProjectRole(RoleViewer))
}

func TestValidProjectRole_Invalid(t *testing.T) {
	assert.False(t, ValidProjectRole("admin"))
	assert.False(t, ValidProjectRole(""))
	assert.False(t, ValidProjectRole("superuser"))
}
