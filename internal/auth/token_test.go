package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssueAndValidateAccessToken(t *testing.T) {
	ts := NewTokenService("test-secret-at-least-32-chars!!", 15*time.Minute, 7*24*time.Hour)

	token, err := ts.IssueAccessToken("user-123")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	userID, err := ts.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", userID)
}

func TestIssueAndValidateRefreshToken(t *testing.T) {
	ts := NewTokenService("test-secret-at-least-32-chars!!", 15*time.Minute, 7*24*time.Hour)

	token, err := ts.IssueRefreshToken("user-456")
	require.NoError(t, err)

	userID, err := ts.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-456", userID)
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	ts1 := NewTokenService("secret-one-at-least-32-chars!!!", 15*time.Minute, 7*24*time.Hour)
	ts2 := NewTokenService("secret-two-at-least-32-chars!!!", 15*time.Minute, 7*24*time.Hour)

	token, err := ts1.IssueAccessToken("user-123")
	require.NoError(t, err)

	_, err = ts2.ValidateToken(token)
	assert.Error(t, err)
}

func TestValidateToken_Expired(t *testing.T) {
	ts := NewTokenService("test-secret-at-least-32-chars!!", -1*time.Second, 7*24*time.Hour)

	token, err := ts.IssueAccessToken("user-123")
	require.NoError(t, err)

	_, err = ts.ValidateToken(token)
	assert.Error(t, err)
}

func TestValidateToken_Garbage(t *testing.T) {
	ts := NewTokenService("test-secret-at-least-32-chars!!", 15*time.Minute, 7*24*time.Hour)

	_, err := ts.ValidateToken("not.a.valid.token")
	assert.Error(t, err)
}
