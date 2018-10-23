package archetype

import "fmt"

func reactorName(reactor interface{}) string {
	reactorName := fmt.Sprintf("%T", reactor)
	reactorNameProvider, reactorNameProviderOk := reactor.(ReactorNameProvider)
	if reactorNameProviderOk {
		reactorName = reactorNameProvider.ReactorName()
	}
	return reactorName
}
