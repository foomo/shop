package unique

import (
	"fmt"

	"github.com/bwmarrin/snowflake"
)

/* Package go-shortid guarantees the generation of unique Ids
with zero collisions for 34 years (1/1/2016-1/1/2050) */

//var generator *shortid.Shortid

// GetNewID returns a new unique identifier string // uses "github.com/ventu-io/go-shortid"
// func GetNewID() string {
// 	var seed uint64 = 3214
// 	if generator == nil {
// 		newGenerator, err := shortid.New(1, shortid.DefaultABC, seed)
// 		generator = newGenerator
// 		if err != nil {
// 			// The Shop can no longer work without this, therfore panic.
// 			panic(err)
// 		}
// 	}

// 	id, err := generator.Generate()
// 	if err != nil {
// 		// The Shop can no longer work without this, therefore panic.
// 		panic(err)
// 	}
// 	return id
// }

var node *snowflake.Node

func GetNewID() string {
	if node == nil {
		var createNodeErr error
		node, createNodeErr = snowflake.NewNode(1)
		if createNodeErr != nil {
			fmt.Println(createNodeErr)
			return ""
		}
	}
	// Generate a snowflake ID.
	id := node.Generate()
	return fmt.Sprintf("%s", id)
}
