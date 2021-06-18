package services

import (
	"github.com/Spruik/libre-common/common/core/ports"
)

type tagChangeHandlerFactoryService struct {
	port ports.TagChangeHandlerFactoryPort
}

func NewTagChangeHandlerFactoryService(port ports.TagChangeHandlerFactoryPort) *tagChangeHandlerFactoryService {
	var ret = tagChangeHandlerFactoryService{}
	ret.port = port
	return &ret
}

var tagChangeHandlerFactoryServiceInstance *tagChangeHandlerFactoryService = nil

func SetTagChangeHandlerFactoryServiceInstance(inst *tagChangeHandlerFactoryService) {
	tagChangeHandlerFactoryServiceInstance = inst
}
func GetTagChangeHandlerFactoryServiceInstance() *tagChangeHandlerFactoryService {
	return tagChangeHandlerFactoryServiceInstance
}

func (s *tagChangeHandlerFactoryService) CreateHandlerInstance(key string, mgdEq *ports.ManagedEquipmentPort) ports.TagChangeHandlerPort {
	return s.port.CreateHandlerInstance(key, mgdEq)
}
