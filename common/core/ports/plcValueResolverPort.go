package ports

import "github.com/Spruik/libre-common/common/core/domain"

//The PlcValueResolverIF interface describes the functions to be provided by any Tag value resolver
type PlcValueResolverPort interface {
	//ResolvePlcValue returns the internal value of a given property given an external value
	ResolvePlcValue(propName domain.StdMessageStruct, plcValue string) (string, error)

	//ResolveStdValue returns the external value of a given property given an internal value
	ResolveStdValue(propName domain.StdMessageStruct, stdValue string) (string, error)
}
