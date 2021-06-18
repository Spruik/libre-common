package utilities

import (
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-common/common/core/services"
	"github.com/Spruik/libre-configuration"
	"github.com/Spruik/libre-logging"
)

type equipmentCacheDefault struct {
	//inherit logging functions
	libreLogger.LoggingEnabler

	//inherit config functions
	libreConfig.ConfigurationEnabler

	dataStore  ports.LibreDataStorePort
	finderIF   ports.EquipmentFinderPort
	nameCache  map[string]*ports.ManagedEquipmentPort
	idCache    map[string]*ports.ManagedEquipmentPort
	cofigLevel int
}

func NewEquipmentCacheDefault(storeIF ports.LibreDataStorePort, finderIF ports.EquipmentFinderPort) *equipmentCacheDefault {
	s := equipmentCacheDefault{
		dataStore:  storeIF,
		finderIF:   finderIF,
		nameCache:  map[string]*ports.ManagedEquipmentPort{},
		idCache:    map[string]*ports.ManagedEquipmentPort{},
		cofigLevel: 0,
	}
	s.SetLoggerConfigHook("EQCACHE")
	s.SetConfigCategory("equipmentCacheDefault")
	return &s
}

func (s *equipmentCacheDefault) RefreshCache() {
	eqList, err := s.finderIF.FindEquipment()
	if err == nil {
		for _, eq := range eqList {
			//update the caches
			existingNameItem := s.nameCache[eq.Name]
			existingIdItem := s.idCache[eq.Id]
			if existingNameItem == nil || existingIdItem == nil {
				//add a new entry
				newEntry := services.GetManagedEquipmentFactoryServiceInstance().GetNewInstance(eq)
				s.nameCache[eq.Name] = &newEntry
				s.idCache[eq.Id] = &newEntry
			} else {
				(*existingNameItem).SetConfigLevel(s.cofigLevel)
			}
		}
	} else {
		s.LogErrorf("FAILED IN CACHE REFRESH: %s", err)
	}
}

func (s *equipmentCacheDefault) GetEquipmentCacheList() *map[string]*ports.ManagedEquipmentPort {
	return &s.nameCache
}

func (s *equipmentCacheDefault) GetCachedEquipmentItem(equipName string) *ports.ManagedEquipmentPort {
	return s.nameCache[equipName]
}

func (s *equipmentCacheDefault) GetCachedEquipmentItemById(equipId string) *ports.ManagedEquipmentPort {
	return s.idCache[equipId]
}
