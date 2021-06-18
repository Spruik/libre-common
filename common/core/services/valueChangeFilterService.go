package services

import (
	"github.com/Spruik/libre-common/common/core/ports"
)

type valueChangeFilterFactoryService struct {
	port ports.ValueChangeFilterFactoryPort
}

var valueChangeFilterFactoryServiceInstance *valueChangeFilterFactoryService = nil

func SetValueChangeFilterFactoryServiceInstance(inst *valueChangeFilterFactoryService) {
	valueChangeFilterFactoryServiceInstance = inst
}
func GetValueChangeFilterFactoryServiceInstance() *valueChangeFilterFactoryService {
	return valueChangeFilterFactoryServiceInstance
}

func NewValueChangeFilterFactoryService(port ports.ValueChangeFilterFactoryPort) *valueChangeFilterFactoryService {
	var ret = valueChangeFilterFactoryService{}
	ret.port = port
	return &ret
}

func (s *valueChangeFilterFactoryService) CreateFilterInstance(key string, mgdEq *ports.ManagedEquipmentPort) ports.ValueChangeFilterPort {
	return s.port.CreateFilterInstance(key, mgdEq)
}
