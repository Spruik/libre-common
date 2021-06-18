package ports

//The PlcEqNameResolverPort interface describes the functions to be provided by any Equipment name resolver
type PlcEqNameResolverPort interface {
	//ResolvePlcName returns the standard (internal) name for an Equipment given the external name
	ResolvePlcEqName(plcName string) (string, error)

	//ResolveStdName returns the external (PLC) name for an Equipment given the interanl name
	ResolveStdEqName(stdName string) (string, error)
}
