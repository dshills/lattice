package domain

import "fmt"

// State represents the lifecycle stage of a WorkItem.
type State string

const (
	NotDone    State = "NotDone"
	InProgress State = "InProgress"
	Completed  State = "Completed"
)

var validStates = map[State]bool{
	NotDone:    true,
	InProgress: true,
	Completed:  true,
}

// ValidState returns true if the given state is one of the three allowed values.
func ValidState(s State) bool {
	return validStates[s]
}

// forwardTransitions defines the allowed single-step forward state transitions.
var forwardTransitions = map[State]State{
	NotDone:    InProgress,
	InProgress: Completed,
}

// stateOrder maps states to their ordinal position for backward detection.
var stateOrder = map[State]int{NotDone: 0, InProgress: 1, Completed: 2}

// ValidateTransition checks whether a state transition is allowed.
// Only single-step forward transitions are permitted (NotDone→InProgress, InProgress→Completed).
// Skipping states (e.g., NotDone→Completed) is not allowed.
// Backward transitions require override=true and isAdmin=true.
// The override field is ignored for forward transitions.
func ValidateTransition(current, next State, override bool, isAdmin bool) error {
	if !ValidState(current) {
		return fmt.Errorf("%w: invalid current state %q", ErrInvalidInput, current)
	}
	if !ValidState(next) {
		return fmt.Errorf("%w: invalid next state %q", ErrInvalidInput, next)
	}
	if current == next {
		return nil
	}

	// Check if this is a valid single-step forward transition.
	if allowed, ok := forwardTransitions[current]; ok && allowed == next {
		return nil
	}

	// This is a backward or skipped transition.
	if !override {
		return fmt.Errorf("%w: transition from %s to %s is not allowed", ErrInvalidTransition, current, next)
	}
	if !isAdmin {
		return fmt.Errorf("%w: override requires admin role", ErrForbidden)
	}

	// Backward transition with admin override.
	if stateOrder[current] > stateOrder[next] {
		return nil
	}

	return fmt.Errorf("%w: transition from %s to %s is not allowed", ErrInvalidTransition, current, next)
}
