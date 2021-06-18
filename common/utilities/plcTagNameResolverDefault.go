package utilities

import "github.com/Spruik/libre-common/common/core/domain"

type plcTagNameResolverDefault struct {
}

//NewPlcTagNameResolverDefault builds a new instance of the default (do nothing) tag name/property resolver
func NewPlcTagNameResolverDefault() *plcTagNameResolverDefault {
	return &plcTagNameResolverDefault{}
}

func (s *plcTagNameResolverDefault) ResolvePlcTagName(plcName string) (string, error) {
	//default is to just pass the same name
	return plcName, nil
}

func (s *plcTagNameResolverDefault) ResolveStdTagName(stdName domain.StdMessageStruct) (string, error) {
	//default is to just pass the same name
	return stdName.ItemName, nil
}
