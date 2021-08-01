package services

import (
	"github.com/Spruik/libre-common/common/core/ports"
)

type equipmentCacheService struct {
	port ports.EquipmentCachePort
}

func NewEquipmentCacheService(port ports.EquipmentCachePort) *equipmentCacheService {
	var ret = equipmentCacheService{}
	ret.port = port
	return &ret
}

var equipmentCacheServiceInstance *equipmentCacheService = nil

func SetEquipmentCacheServiceInstance(inst *equipmentCacheService) {
	equipmentCacheServiceInstance = inst
}
func GetEquipmentCacheServiceInstance() *equipmentCacheService {
	return equipmentCacheServiceInstance
}

func (s *equipmentCacheService) RefreshCache() {
	s.port.RefreshCache()
}

func (s *equipmentCacheService) GetEquipmentCacheList() *map[string]*ports.ManagedEquipmentPort {
	return s.port.GetEquipmentCacheList()
}

func (s *equipmentCacheService) GetCachedEquipmentItem(equipName string) *ports.ManagedEquipmentPort {
	return s.port.GetCachedEquipmentItem(equipName)
}

func (s *equipmentCacheService) GetCachedEquipmentItemById(equipId string) *ports.ManagedEquipmentPort {
	return s.port.GetCachedEquipmentItemById(equipId)
}
func (s *equipmentCacheService) SetEquipmentChangeNoticeFunction(handlingFxn func(notice ports.EquipmentCacheChangeNotice)) {
	s.port.SetEquipmentChangeNoticeFunction(handlingFxn)
}

func (s *equipmentCacheService) StartMonitoring() {
	s.port.StartMonitoring()
}

func (s *equipmentCacheService) StopMonitoring() {
	s.port.StopMonitoring()

}
