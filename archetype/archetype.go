package archetype

import (
	"fmt"
	"runtime"
)

func reactorName(reactor interface{}) string {
	// If there isn't one, who is the caller?
	if reactor == nil {
		pc, _, _, _ := runtime.Caller(1)
		reactor = runtime.FuncForPC(pc)
	}

	reactorName := fmt.Sprintf("%T", reactor)
	reactorNameProvider, reactorNameProviderOk := reactor.(ReactorNameProvider)
	if reactorNameProviderOk {
		reactorName = reactorNameProvider.ReactorName()
	}
	return reactorName
}
