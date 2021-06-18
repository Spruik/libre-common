package services

import (
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
)

type managedEquipmentFactoryService struct {
	port ports.ManagedEquipmentFactoryIF
}

func NewManagedEquipmentFactoryService(port ports.ManagedEquipmentFactoryIF) *managedEquipmentFactoryService {
	var ret = managedEquipmentFactoryService{}
	ret.port = port
	return &ret
}

var managedEquipmentFactoryServiceInstance *managedEquipmentFactoryService = nil

func SetManagedEquipmentFactoryServiceInstance(inst *managedEquipmentFactoryService) {
	managedEquipmentFactoryServiceInstance = inst
}
func GetManagedEquipmentFactoryServiceInstance() *managedEquipmentFactoryService {
	return managedEquipmentFactoryServiceInstance
}

func (s *managedEquipmentFactoryService) GetNewInstance(eqInst domain.Equipment) ports.ManagedEquipmentPort {
	return s.port.GetNewInstance(eqInst)
}
