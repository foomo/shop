package persistence

import (
	"testing"

	"gopkg.in/mgo.v2"
)

type Foo struct {
	FirstName string
	LastName  string
}

func TestPersistenceIndex(t *testing.T) {
	index := mgo.Index{
		Key:        []string{"firstname", "lastname"},
		Unique:     true,
		Background: true,
	}
	p, err := NewPersistorWithIndex("mongodb://dockerhost/test", "testindex", index)
	if err != nil {
		t.Fatal(err)
	}
	err = p.GetCollection().EnsureIndex(index)
	if err != nil {
		t.Fatal(err)
	}
	err = p.GetCollection().Insert(&Foo{
		FirstName: "Foo",
		LastName:  "Bar",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = p.GetCollection().Insert(&Foo{
		FirstName: "Flo",
		LastName:  "Bar",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = p.GetCollection().Insert(&Foo{
		FirstName: "Flo",
		LastName:  "Bar",
	})
	if err == nil {
		t.Fail()
		t.Log("Did not expected that one to work!")
	}
	if err != nil {
		t.Log("Error: " + err.Error())
	}
}
