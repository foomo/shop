package unique

import (
	"fmt"

	"github.com/bwmarrin/snowflake"
	"github.com/ventu-io/go-shortid"
)

var node *snowflake.Node
var generator *shortid.Shortid

func GetNewID() string {
	//return GetNewIDSnowFlake() // SnowFlake is default
	id := GetNewIDSnowFlake() // SnowFlake is default
	return id
}

// GetNewIDSnowFlake returns a new unique identifier string
// Generates numeric strings
func GetNewIDSnowFlake() string {
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

// GetNewIDShortId returns a new unique identifier string
// Package go-shortid guarantees the generation of unique Ids
// with zero collisions for 34 years (1/1/2016-1/1/2050) */
func GetNewIDShortID() string {
	var seed uint64 = 3214
	if generator == nil {
		newGenerator, err := shortid.New(1, shortid.DEFAULT_ABC, seed)
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
