package state

import (
	"log"
	"testing"
)

const (
	StateType = "ExampleStates"
	State1    = "state1"
	State2    = "state2"
	State3    = "state3"
	State4    = "state4"
)

var transitions = map[string][]string{
	State1: []string{State2},
	State2: []string{WILDCARD},
	State3: []string{State4},
	State4: []string{},
}

// blueprints for possible states
var blueprints = map[string]BluePrint{
	State1: BluePrint{
		Type:        StateType,
		Key:         State1,
		Description: "I am state one",
		Initial:     true,
	},
	State2: BluePrint{
		Type:        StateType,
		Key:         State2,
		Description: "I am state two",
		Initial:     false,
	},
	State3: BluePrint{
		Type:        StateType,
		Key:         State3,
		Description: "I am state three",
		Initial:     false,
	},
	State4: BluePrint{
		Type:        StateType,
		Key:         State4,
		Description: "I am state four",
		Initial:     false,
	},
}

var stateMachine = StateMachine{
	InitialState: State1,
	Transitions:  transitions,
	BluePrints:   blueprints,
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

	// Go from state 2 to state 4. This should be possible because of WILDCARD
	state, err = stateMachine.TransitionToState(state, State4)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Current State: ", state.Key)

}
