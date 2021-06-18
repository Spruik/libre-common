package services

import (
	"github.com/Spruik/libre-common/common/core/ports"
)

type plcEqNameResolverService struct {
	port ports.PlcEqNameResolverPort
}

func NewPlcEqNameResolverService(port ports.PlcEqNameResolverPort) *plcEqNameResolverService {
	var ret plcEqNameResolverService
	ret.port = port
	return &ret
}

var plcEqNameResolverServiceInstance *plcEqNameResolverService = nil

func SetPlcEqNameResolverServiceInstance(inst *plcEqNameResolverService) {
	plcEqNameResolverServiceInstance = inst
}
func GetPlcEqNameResolverServiceInstance() *plcEqNameResolverService {
	return plcEqNameResolverServiceInstance
}

func (s *plcEqNameResolverService) ResolvePlcEqName(plcName string) (string, error) {
	ret, err := s.port.ResolvePlcEqName(plcName)
	if err == nil && ret == "" {
		ret = plcName
	}
	return ret, err
}

func (s *plcEqNameResolverService) ResolveStdEqName(stdName string) (string, error) {
	ret, err := s.port.ResolveStdEqName(stdName)
	if err == nil && ret == "" {
		ret = stdName
	}
	return ret, err
}
