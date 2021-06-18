package ports

import "github.com/Spruik/libre-common/common/core/domain"

//The PlcTagNameResolverPort interface describes the functions to be provided by any Tag name resolver
type PlcTagNameResolverPort interface {

	//ResolvePlcName returns the standard name for a Tag given the external name
	ResolvePlcTagName(plcName string) (string, error)

	//ResolveStdName returns the external name for a Tag given the internal name - since internally, names are equipment dependent, we pass a TagInfo struct
	ResolveStdTagName(stdName domain.StdMessageStruct) (string, error)
}
