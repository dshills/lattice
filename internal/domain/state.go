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

// ValidateTransition checks whether a state transition is allowed.
// Any transition between valid states is permitted when override=true.
// Without override, only single-step forward transitions are allowed
// (NotDone→InProgress, InProgress→Completed).
func ValidateTransition(current, next State, override bool) error {
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

	// Any other transition requires override.
	if !override {
		return fmt.Errorf("%w: transition from %s to %s is not allowed", ErrInvalidTransition, current, next)
	}

	return nil
}
