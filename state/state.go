package state

import (
	"errors"
	"log"
	"time"
)

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
	Initial        bool
	CreatedAt      time.Time
	LastModifiedAt time.Time
}

type StateFactoryFunc func(key string) *State

type StateMachine struct {
	InitialState string // key of initial state
	Transitions  map[string][]string
	StateFactory StateFactoryFunc
}

func (sm *StateMachine) TransitionToState(currentState *State, targetState string) (*State, error) {
	return sm.transitionToState(currentState, targetState, false)
}
func (sm *StateMachine) ForceTransitionToState(currentState *State, targetState string) (*State, error) {
	return sm.transitionToState(currentState, targetState, true)
}

// TransitionToState returns the target state if this transition is possible.
// If force, target state is returned in any case.
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
		if targetState == transition {
			return sm.StateFactory(targetState), nil
		}
	}
	return currentState, errors.New("Transition from " + currentState.Key + " to " + targetState + " not possible.")

}
func (sm *StateMachine) GetInitialState() *State {
	return sm.StateFactory(sm.InitialState)
}
