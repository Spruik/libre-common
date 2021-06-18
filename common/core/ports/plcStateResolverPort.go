package ports

import "github.com/Spruik/libre-common/common/core/domain"

//The PlcStateResolverPort interface describes the functions to be provided by any state (as in state-machine) resolver
type PlcStateResolverPort interface {
	//ResolvePlcState returns the internal state name used by the given property given an external value
	ResolvePlcState(propName domain.Property, plcState string) (string, error)

	//ResolveStdState returns the external state name used by a given property given an internal value
	ResolveStdState(propName domain.Property, stdState string) (string, error)
}
