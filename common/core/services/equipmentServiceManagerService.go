package services

import (
	"github.com/Spruik/libre-common/common/core/ports"
	"sync"
)

type equipmentServiceManagerService struct {
	port ports.EquipmentServiceManagerIF
}

func NewEquipmentServiceManagerService(port ports.EquipmentServiceManagerIF) *equipmentServiceManagerService {
	var ret equipmentServiceManagerService
	ret.port = port
	return &ret
}

var equipmentServiceManagerServiceInstance *equipmentServiceManagerService = nil

func SetEquipmentServiceManagerServiceInstance(inst *equipmentServiceManagerService) {
	equipmentServiceManagerServiceInstance = inst
}
func GetEquipmentServiceManagerServiceInstance() *equipmentServiceManagerService {
	return equipmentServiceManagerServiceInstance
}

func (s *equipmentServiceManagerService) Initialize() error {
	return s.port.Initialize()
}

func (s *equipmentServiceManagerService) Start(wg *sync.WaitGroup) error {
	return s.port.Start(wg)
}

func (s *equipmentServiceManagerService) Shutdown() error {
	return s.port.Shutdown()
}
