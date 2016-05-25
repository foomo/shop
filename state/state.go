package state

import (
	"errors"
	"log"
	"time"
)

//------------------------------------------------------------------
// ~ CONSTANTS
//------------------------------------------------------------------

const WILDCARD = "*" // a state with this target can transition to any other state

//------------------------------------------------------------------
// ~ PUBLIC TYPES
//------------------------------------------------------------------

type BluePrint struct {
	Type        string
	Key         string
	Description string
	Initial     bool
}
type State struct {
	Type           string
	Key            string
	Description    string
	CreatedAt      time.Time
	LastModifiedAt time.Time
}

type StateFactoryFunc func(key string) *State

type StateMachine struct {
	InitialState string // key of initial state
	Transitions  map[string][]string
	StateFactory StateFactoryFunc
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// GetInitialState returns the initial state
func (sm *StateMachine) GetInitialState() *State {
	return sm.StateFactory(sm.InitialState)
}

// TransitionToState returns target state if possible, else current state
func (sm *StateMachine) TransitionToState(currentState *State, targetState string) (*State, error) {
	return sm.transitionToState(currentState, targetState, false)
}

// ForceTransitionToState returns target state whether the transition is possible or not
func (sm *StateMachine) ForceTransitionToState(currentState *State, targetState string) (*State, error) {
	return sm.transitionToState(currentState, targetState, true)
}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------

// TransitionToState returns the target state if this transition is possible, else current state.
// If force, target state is returned whether the transition is possible or not
func (sm *StateMachine) transitionToState(currentState *State, targetState string, force bool) (*State, error) {
	if force {
		return sm.StateFactory(targetState), nil
	}
	// Get the possible transitions for currentState
	transitions, ok := sm.Transitions[currentState.Key]
	log.Println(transitions)
	if !ok {
		return currentState, errors.New(targetState + " is not a valid state!")
	}
	// Check if targetState is a possible target state
	for _, transition := range transitions {
		if targetState == transition || transition == WILDCARD {
			return sm.StateFactory(targetState), nil
		}
	}
	return currentState, errors.New("Transition from " + currentState.Key + " to " + targetState + " not possible.")

}
