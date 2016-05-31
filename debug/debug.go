package debug

import "log"

const VERBOSE = false

// Log A logger which can be switched on and off globally with VERBOSE
func Log(i ...interface{}) {
	if VERBOSE {
		log.Println("[DEBUG]", i)
	}
}
