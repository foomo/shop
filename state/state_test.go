package state

import (
	"log"
	"testing"

	"github.com/foomo/shop/utils"
)

const (
	StateType = "ExampleStates"
	State1    = "state1"
	State2    = "state2"
	State3    = "state3"
	State4    = "state4"
)

var stateDescriptions = map[string]string{
	State1: "I am state one",
	State2: "I am state two",
	State3: "I am state three",
	State4: "I am state four",
}

var transitions = map[string][]string{
	State1: []string{State2},
	State2: []string{State3},
	State3: []string{State4},
	State4: []string{},
}

// blueprints for possible states
var blueprints = map[string]BluePrint{
	State1: BluePrint{
		Type:        StateType,
		Key:         State1,
		Description: stateDescriptions[State1],
		Initial:     true,
	},
	State2: BluePrint{
		Type:        StateType,
		Key:         State2,
		Description: stateDescriptions[State2],
		Initial:     false,
	},
	State3: BluePrint{
		Type:        StateType,
		Key:         State3,
		Description: stateDescriptions[State3],
		Initial:     false,
	},
	State4: BluePrint{
		Type:        StateType,
		Key:         State4,
		Description: stateDescriptions[State4],
		Initial:     false,
	},
}

func stateFactory(key string) *State {

	blueprint := blueprints[key]

	return &State{
		CreatedAt:      utils.TimeNow(),
		LastModifiedAt: utils.TimeNow(),
		Type:           blueprint.Type,
		Key:            blueprint.Key,
		Description:    blueprint.Description,
	}
}

var stateMachine = StateMachine{
	InitialState: State1,
	Transitions:  transitions,
	StateFactory: stateFactory,
}

func TestStates(t *testing.T) {
	state := stateMachine.GetInitialState()
	log.Println("Current State: ", state.Key)

	// Go to next state
	state, err := stateMachine.TransitionToState(state, State2)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Current State: ", state.Key)
	// Go to next state
	state, err = stateMachine.TransitionToState(state, State3)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Current State: ", state.Key)

	// Go to previous state. This should fail
	state, err = stateMachine.TransitionToState(state, State2)
	if err == nil {
		t.Fail()
	}
	log.Println(err.Error())
	log.Println("Current State: ", state.Key)

	// Force transition to previous state
	state, err = stateMachine.ForceTransitionToState(state, State2)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Current State: ", state.Key)

}
