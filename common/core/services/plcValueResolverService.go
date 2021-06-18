package services

import (
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
)

type plcValueResolverService struct {
	port ports.PlcValueResolverPort
}

func NewValueResolverService(port ports.PlcValueResolverPort) *plcValueResolverService {
	var ret plcValueResolverService
	ret.port = port
	return &ret
}

var plcValueResolverServiceInstance *plcValueResolverService = nil

func SetPlcValueResolverServiceInstance(inst *plcValueResolverService) {
	plcValueResolverServiceInstance = inst
}
func GetPlcValueResolverServiceInstance() *plcValueResolverService {
	return plcValueResolverServiceInstance
}

func (s *plcValueResolverService) ResolvePlcValue(propName domain.StdMessageStruct, plcValue string) (string, error) {
	ret, err := s.port.ResolvePlcValue(propName, plcValue)
	if err == nil && ret == "" {
		ret = plcValue
	}
	return ret, err
}

func (s *plcValueResolverService) ResolveStdValue(propName domain.StdMessageStruct, stdValue string) (string, error) {
	ret, err := s.port.ResolveStdValue(propName, stdValue)
	if err == nil && ret == "" {
		ret = stdValue
	}
	return ret, err
}
