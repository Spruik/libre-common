package services

import (
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
)

type plcTagNameResolverService struct {
	port ports.PlcTagNameResolverPort
}

func NewPlcTagNameResolverService(port ports.PlcTagNameResolverPort) *plcTagNameResolverService {
	return &plcTagNameResolverService{
		port: port,
	}
}

var plcTagNameResolverServiceInstance *plcTagNameResolverService = nil

func SetPlcTagNameResolverServiceInstance(inst *plcTagNameResolverService) {
	plcTagNameResolverServiceInstance = inst
}
func GetPlcTagNameResolverServiceInstance() *plcTagNameResolverService {
	return plcTagNameResolverServiceInstance
}

func (s *plcTagNameResolverService) ResolvePlcTagName(plcName string, eqName string) (string, error) {
	return s.port.ResolvePlcTagName(plcName, eqName)
}

func (s *plcTagNameResolverService) ResolveStdTagName(stdName domain.StdMessageStruct) (string, error) {
	return s.port.ResolveStdTagName(stdName)
}
