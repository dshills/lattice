package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("correcthorse")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "correcthorse", hash)

	assert.NoError(t, CheckPassword(hash, "correcthorse"))
}

func TestCheckPassword_Wrong(t *testing.T) {
	hash, err := HashPassword("correcthorse")
	require.NoError(t, err)

	assert.Error(t, CheckPassword(hash, "wrongpassword"))
}

func TestHashPassword_Unique(t *testing.T) {
	h1, err := HashPassword("samepassword")
	require.NoError(t, err)
	h2, err := HashPassword("samepassword")
	require.NoError(t, err)

	assert.NotEqual(t, h1, h2, "bcrypt should produce different hashes for same input")
}
