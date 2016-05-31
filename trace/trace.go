package trace

import (
	"runtime"

	"github.com/foomo/shop/utils"
)

type Callers []string

func (c Callers) String() {
	s := ""
	for i, caller := range c {
		s += caller + utils.IteString(i != len(c)-1, "\n", "")
	}
}

// WhoCalledMe returns a string slice containing all calling functions of the stack trace
// Functions until depth of start are skipped
func WhoCalledMe(start int) Callers {
	var callers Callers
	i := start
	for true {
		caller := whoCalledMeSkip(i)
		if caller != "n/a" {
			callers = append(callers, caller)
			i++
			continue
		}
		break
	}
	return callers
}

func whoCalledMeSkip(skip int) string {
	// we get the callers as uintptrs - but we just need 1
	fpcs := make([]uintptr, 1)

	// skip 3 levels to get to the caller of whoever called Caller()
	n := runtime.Callers(skip, fpcs)
	if n == 0 {
		return "n/a" // proper error her would be better
	}

	// get the info of the actual function that's in the pointer
	fun := runtime.FuncForPC(fpcs[0] - 1)
	if fun == nil {
		return "n/a"
	}

	// return its name
	return fun.Name()
}
