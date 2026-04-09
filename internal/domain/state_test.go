package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateTransition_ForwardAllowed(t *testing.T) {
	tests := []struct {
		from, to State
	}{
		{NotDone, InProgress},
		{InProgress, Completed},
	}
	for _, tt := range tests {
		assert.NoError(t, ValidateTransition(tt.from, tt.to, false),
			"%s → %s should be allowed", tt.from, tt.to)
	}
}

func TestValidateTransition_SameState(t *testing.T) {
	for _, s := range []State{NotDone, InProgress, Completed} {
		assert.NoError(t, ValidateTransition(s, s, false))
	}
}

func TestValidateTransition_SkipForwardNotAllowed(t *testing.T) {
	err := ValidateTransition(NotDone, Completed, false)
	assert.ErrorIs(t, err, ErrInvalidTransition)
}

func TestValidateTransition_BackwardDeniedWithoutOverride(t *testing.T) {
	tests := []struct {
		from, to State
	}{
		{Completed, InProgress},
		{Completed, NotDone},
		{InProgress, NotDone},
	}
	for _, tt := range tests {
		err := ValidateTransition(tt.from, tt.to, false)
		assert.ErrorIs(t, err, ErrInvalidTransition,
			"%s → %s should be denied without override", tt.from, tt.to)
	}
}

func TestValidateTransition_BackwardAllowedWithOverride(t *testing.T) {
	// Any user can move backward with override=true.
	err := ValidateTransition(Completed, InProgress, true)
	assert.NoError(t, err)
}

func TestValidateTransition_BackwardAllowedWithOverride_AllDirections(t *testing.T) {
	tests := []struct {
		from, to State
	}{
		{Completed, InProgress},
		{Completed, NotDone},
		{InProgress, NotDone},
	}
	for _, tt := range tests {
		assert.NoError(t, ValidateTransition(tt.from, tt.to, true),
			"%s → %s should be allowed with override", tt.from, tt.to)
	}
}

func TestValidateTransition_InvalidStates(t *testing.T) {
	err := ValidateTransition("invalid", NotDone, false)
	assert.True(t, errors.Is(err, ErrInvalidInput))

	err = ValidateTransition(NotDone, "bogus", false)
	assert.True(t, errors.Is(err, ErrInvalidInput))
}

func TestValidateTransition_OverrideIgnoredForForward(t *testing.T) {
	// override=true on a forward transition should still succeed (ignored)
	assert.NoError(t, ValidateTransition(NotDone, InProgress, true))
}
