package unique

import "github.com/ventu-io/go-shortid"

/* Package go-shortid guarantees the generation of unique Ids
with zero collisions for 34 years (1/1/2016-1/1/2050) */

var generator *shortid.Shortid

// GetNewID returns a new unique identifier string
func GetNewID() string {
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
