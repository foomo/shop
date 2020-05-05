package unique

import (
	"github.com/ventu-io/go-shortid"

	uuid "github.com/satori/go.uuid"
)

var generator *shortid.Shortid

// GetNewID generates a proper V4 UUID
func GetNewID() string {
	return uuid.NewV4().String()
}

// GetNewIDShortId returns a new unique identifier string
// Package go-shortid guarantees the generation of unique Ids
// with zero collisions for 34 years (1/1/2016-1/1/2050) */
func GetNewIDShortID() string {
	var seed uint64 = 3214
	if generator == nil {
		newGenerator, err := shortid.New(1, shortid.DefaultABC, seed)
		generator = newGenerator
		if err != nil {
			// The Shop can no longer work without this, therfore panic.
			panic(err)
		}
	}

	id, err := generator.Generate()
	if err != nil {
		// The Shop can no longer work without this, therefore panic.
		panic(err)
	}
	return id
}
