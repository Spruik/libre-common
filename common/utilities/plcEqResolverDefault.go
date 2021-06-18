package utilities

import "github.com/Spruik/libre-common/common/core/domain"

type plcEqNameResolverDefault struct {
}

//NewPlcEqNameResolverDefault builds a new instance of the default (do nothing) equipment name resolver
func NewPlcEqNameResolverDefault() *plcEqNameResolverDefault {
	return &plcEqNameResolverDefault{}
}

func (s *plcEqNameResolverDefault) ResolvePlcEqName(plcName string) (string, error) {
	//default is to just pass the same name
	return plcName, nil
}

func (s *plcEqNameResolverDefault) ResolveStdEqName(stdName domain.Equipment) (string, error) {
	//default is to just pass the same name
	return stdName.Name, nil
}
