package state

import (
	"errors"
	"time"

	"github.com/foomo/shop/utils"
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

type StateMachine struct {
	InitialState string // key of initial state
	Transitions  map[string][]string
	BluePrints   map[string]BluePrint
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// GetInitialState returns the initial state
func (sm *StateMachine) GetInitialState() *State {
	initialState, _ := sm.stateFactory(sm.InitialState)
	return initialState
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
		return sm.stateFactory(targetState)
	}
	// Get the possible transitions for currentState
	transitions, ok := sm.Transitions[currentState.Key]
	if !ok {
		return currentState, errors.New("No transitions defined for " + targetState)
	}
	// Check if targetState is a possible target state
	for _, transition := range transitions {
		if targetState == transition || transition == WILDCARD {
			return sm.stateFactory(targetState)
		}
	}
	return currentState, errors.New("Transition from " + currentState.Key + " to " + targetState + " not possible.")

}

// return target State
func (sm *StateMachine) stateFactory(key string) (*State, error) {
	blueprint, ok := sm.BluePrints[key]
	if !ok {
		return nil, errors.New(key + " is not a valid state.")
	}

	return &State{
		CreatedAt:      utils.TimeNow(),
		LastModifiedAt: utils.TimeNow(),
		Type:           blueprint.Type,
		Key:            blueprint.Key,
		Description:    blueprint.Description,
	}, nil
}
