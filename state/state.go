package state

import (
	"errors"
	"log"
	"time"

	"github.com/foomo/shop/utils"
)

//------------------------------------------------------------------
// ~ USAGE

// - Stateful objects have a field of type State.
// - Create a static StateMachine to handle the
//   transitions between states in a safely manner
//------------------------------------------------------------------

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
	Finished       bool
}

type StateMachine struct {
	InitialState string // key of initial state
	Transitions  map[string][]string
	BluePrints   map[string]BluePrint
}

// Types having a *State field, should implement StateMachineContext
type StateMachineContext interface {
	GetStateMachine() *StateMachine
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

func (st *State) SetModified() {
	st.LastModifiedAt = utils.TimeNow()
}

// GetInitialState returns the initial state
func (sm *StateMachine) GetInitialState() *State {
	initialState, _ := sm.stateFactory(sm.InitialState)
	return initialState
}

// TransitionToState if transition is possible, sets currentState to target state
func (sm *StateMachine) TransitionToState(currentState *State, targetState string) error {
	return sm.transitionToState(currentState, targetState, false)
}

// ForceTransitionToState sets currentState to target state whether the transition is possible or not
func (sm *StateMachine) ForceTransitionToState(currentState *State, targetState string) error {
	return sm.transitionToState(currentState, targetState, true)
}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------

// TransitionToState if transition is possible, sets currentState to target state
// If force, target state is returned whether the transition is possible or not
func (sm *StateMachine) transitionToState(currentState *State, targetState string, force bool) error {

	if force {
		if currentState == nil {
			currentState = sm.GetInitialState() // from InitialState we can force go to any other State
		}
		state, err := sm.stateFactory(targetState)
		if err != nil {
			return err
		}
		*currentState = *state
		log.Println("force transitionToState() New current State: ", currentState.Key)
		return nil
	}
	if currentState == nil {
		return errors.New("StateMachineError: Current State is nil")
	}
	// Get the possible transitions for currentState
	transitions, ok := sm.Transitions[currentState.Key]
	if !ok {
		return errors.New("StateMachineError: No transitions defined for " + currentState.Key)
	}

	// Check if targetState is a possible target state
	for _, transition := range transitions {
		if targetState == transition || transition == WILDCARD {
			state, err := sm.stateFactory(targetState)
			if err != nil {
				return err
			}
			*currentState = *state
			log.Println("transitionToState() New current State: ", currentState.Key)
			return nil
		}
	}
	return errors.New("StateMachineError: Transition from " + currentState.Key + " to " + targetState + " not possible.")

}

// return target State
func (sm *StateMachine) stateFactory(key string) (*State, error) {
	blueprint, ok := sm.BluePrints[key]
	if !ok {
		return nil, errors.New("StateMachineError: " + key + " is not a valid state.")
	}

	return &State{
		CreatedAt:      utils.TimeNow(),
		LastModifiedAt: utils.TimeNow(),
		Type:           blueprint.Type,
		Key:            blueprint.Key,
		Description:    blueprint.Description,
	}, nil
}
