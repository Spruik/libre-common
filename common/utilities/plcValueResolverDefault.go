package utilities

import "github.com/Spruik/libre-common/common/core/domain"

type plcValueResolverDefault struct {
}

func NewPlcValueResolverDefault() *plcValueResolverDefault {
	return &plcValueResolverDefault{}
}

func (s *plcValueResolverDefault) ResolvePlcValue(propName domain.StdMessageStruct, plcValue string) (string, error) {
	//default is to just pass the same name
	_ = propName
	return plcValue, nil
}

func (s *plcValueResolverDefault) ResolveStdValue(propName domain.StdMessageStruct, stdValue string) (string, error) {
	//default is to just pass the same name
	_ = propName
	return stdValue, nil
}
