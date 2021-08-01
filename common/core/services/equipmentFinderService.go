package services

import (
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
)

type equipmentFinderService struct {
	port ports.EquipmentFinderPort
}

func NewEquipmentFinderService(port ports.EquipmentFinderPort) *equipmentFinderService {
	var ret = equipmentFinderService{}
	ret.port = port
	return &ret
}

var equipmentFinderServiceInstance *equipmentFinderService = nil

func SetEquipmentFinderServiceInstance(inst *equipmentFinderService) {
	equipmentFinderServiceInstance = inst
}
func GetEquipmentFinderServiceInstance() *equipmentFinderService {
	return equipmentFinderServiceInstance
}
func (s *equipmentFinderService) FindEquipment() ([]domain.Equipment, error) {
	return s.port.FindEquipment()
}
func (s *equipmentFinderService) SubscribeToChanges(notificationChannel chan ports.EquipmentFinderChangeNotice) {
	s.port.SubscribeToChanges(notificationChannel)
}
func (s *equipmentFinderService) UnsubscribeToChanges(notificationChannel chan ports.EquipmentFinderChangeNotice) {
	s.port.UnsubscribeToChanges(notificationChannel)
}
