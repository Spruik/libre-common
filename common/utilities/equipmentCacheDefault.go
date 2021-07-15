package utilities

import (
	"fmt"
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-common/common/core/queries"
	"github.com/Spruik/libre-common/common/core/services"
	"github.com/Spruik/libre-configuration"
	"github.com/Spruik/libre-logging"
	"strconv"
	"time"
)

type equipmentCacheDefault struct {
	//inherit logging functions
	libreLogger.LoggingEnabler

	//inherit config functions
	libreConfig.ConfigurationEnabler

	dataStore          ports.LibreDataStorePort
	finderIF           ports.EquipmentFinderPort
	nameCache          map[string]*ports.ManagedEquipmentPort
	idCache            map[string]*ports.ManagedEquipmentPort
	configLevel        int
	monitorChanges     bool
	monitoringChannel  chan string
	equipmentChangeFxn func(notice ports.EquipmentCacheChangeNotice)
}

func NewEquipmentCacheDefault(storeIF ports.LibreDataStorePort, finderIF ports.EquipmentFinderPort) *equipmentCacheDefault {
	s := equipmentCacheDefault{
		dataStore:         storeIF,
		finderIF:          finderIF,
		nameCache:         map[string]*ports.ManagedEquipmentPort{},
		idCache:           map[string]*ports.ManagedEquipmentPort{},
		configLevel:       0,
		monitoringChannel: nil,
	}
	s.SetLoggerConfigHook("EQCACHE")
	s.SetConfigCategory("equipmentCache")
	monStr, _ := s.GetConfigItemWithDefault("MonitorChanges", "false")
	var err error
	s.monitorChanges, err = strconv.ParseBool(monStr)
	if err != nil {
		panic(fmt.Sprintf("FAILED IN CONFIGURATION SETUP FOR EQUIPMENT CACHE - BAD CONFIG VALUE FOR 'MonitorChanges' :'%s' [%s]", monStr, err))
	}
	s.equipmentChangeFxn = s.defaultEqChangeNoticeHandler
	return &s
}

func (s *equipmentCacheDefault) defaultEqChangeNoticeHandler(notice ports.EquipmentCacheChangeNotice) {
	s.LogWarnf("EquipmentCache noticed an equipment change in the database, but no handler has been set to processes the change.")
}

func (s *equipmentCacheDefault) RefreshCache() {
	s.configLevel++
	eqList, err := s.finderIF.FindEquipment()
	if err == nil {
		for _, eq := range eqList {
			//update the caches
			existingNameItem := s.nameCache[eq.Name]
			existingIdItem := s.idCache[eq.Id]
			if existingNameItem == nil || existingIdItem == nil {
				//add a new entry
				newEntry := services.GetManagedEquipmentFactoryServiceInstance().GetNewInstance(eq)
				newEntry.SetConfigLevel(s.configLevel)
				s.nameCache[eq.Name] = &newEntry
				s.idCache[eq.Id] = &newEntry
				if s.equipmentChangeFxn != nil {
					s.equipmentChangeFxn(ports.EquipmentCacheChangeNotice{
						"ADD",
						eq.Id,
					})
				}
			} else {
				s.updateProperties(eq, existingNameItem)
				(*existingNameItem).SetConfigLevel(s.configLevel)
			}
		}
		//scrub the non-referenced items
		for id, mgdEq := range s.idCache {
			if mgdEq != nil && (*mgdEq).GetConfigLevel() != s.configLevel {
				//equipment has been removed
				delete(s.nameCache, (*mgdEq).GetEquipmentName())
				delete(s.idCache, id)
				if s.equipmentChangeFxn != nil {
					s.equipmentChangeFxn(ports.EquipmentCacheChangeNotice{
						"REMOVE",
						id,
					})
				}
			}
		}
	} else {
		s.LogErrorf("FAILED IN CACHE REFRESH: %s", err)
	}

	//check the monitor settings and crank up monitoring if configured and not already running
	if s.monitorChanges && s.monitoringChannel == nil {
		s.StartMonitoring()
	}
}

func (s *equipmentCacheDefault) updateProperties(eq domain.Equipment, item *ports.ManagedEquipmentPort) {
	//get the properties for the equipment
	txn := s.dataStore.BeginTransaction(false, "eqpropsforcacheupdate")
	defer txn.Dispose()
	eqPropMap, err := queries.GetAllPropertiesForEquipment(txn, eq.Id)
	if err != nil {
		s.LogErrorf("FAILED IN FETCH OF EQUIPMENT PROPERTIES IN CACHE REFRESH! (%s)", eq.Name)
	}
	//add any EQ properties that are missing from the current managed equipment instance
	// note - not trying to remove properties that might have been deleted from the EQ
	currProps := (*item).GetPropertyMap()
	for eqPropName, eqProp := range eqPropMap {
		_, exists := currProps[eqPropName]
		if !exists {
			currProps[eqPropName] = domain.EquipmentPropertyDescriptor{
				Name:             eqPropName,
				DataType:         eqProp.DataType,
				Value:            nil,
				ClassPropertyId:  eqProp.Id,
				EquipmentClassId: eqProp.EquipmentClass.Id,
				LastUpdate:       time.Time{},
			}
		}
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

func (s *equipmentCacheDefault) SetEquipmentChangeNoticeFunction(handlingFxn func(notice ports.EquipmentCacheChangeNotice)) {
	s.equipmentChangeFxn = handlingFxn
}

func (s *equipmentCacheDefault) StartMonitoring() {
	if s.monitoringChannel == nil {
		s.monitoringChannel = make(chan string)
		s.monitorEquipmentChanges()
		select {
		case resp := <-s.monitoringChannel:
			if resp == "RUNNING" {
				s.LogInfo("Equipment Cache monitoring started successfully")
			} else {
				s.LogError("Equipment Cache monitoring start FAILED")
				s.monitoringChannel = nil
			}
		case <-time.After(5 * time.Second):
			s.LogError("Equipment Cache monitoring start TIMED OUT")
			s.monitoringChannel = nil
		}
	} else {
		s.LogErrorf("Monitoring already active when StartMonitoring was called!")
	}
}

func (s *equipmentCacheDefault) StopMonitoring() {
	s.LogDebug("Equipment Cache StopMonitoring begins")
	if s.monitoringChannel != nil {
		s.LogDebug("sending END to monitoring channel")
		s.monitoringChannel <- "END"
		select {
		case resp := <-s.monitoringChannel:
			s.LogDebugf("got response on monitoring channel: %s", resp)
			if resp == "DONE" {
				s.LogInfo("Equipment Cache monitoring stopped successfully")
			} else {
				s.LogError("Equipment Cache monitoring stop FAILED")
				s.monitoringChannel = nil
			}
		case <-time.After(5 * time.Second):
			s.LogError("Equipment Cache monitoring stop TIMED OUT")
			s.monitoringChannel = nil
		}
	} else {
		s.LogError("Monitoring already inactive when StopMonitoring was called!")
	}
	s.LogDebug("Equipment Cache StopMonitoring ends")
}

////////////////////////////////////////////////////////////////////////
func (s *equipmentCacheDefault) monitorEquipmentChanges() {
	finderChannel := make(chan ports.EquipmentFinderChangeNotice)
	go func(c chan ports.EquipmentFinderChangeNotice, m chan string) {
		m <- "RUNNING"
		var haveMsg bool
		var shouldEnd = false
		var msg ports.EquipmentFinderChangeNotice
		var monMsg string
		for !shouldEnd {
			s.LogDebugf("eq cache monitoring routine checking for an admin message")
			select {
			case monMsg = <-m:
				if monMsg == "END" {
					shouldEnd = true
					s.LogInfof("HEY!!!!  got a notice from the using service that I should end!!!!!! %+v", msg)
					s.finderIF.UnsubscribeToChanges(finderChannel)
				}
			default:
			}
			if !shouldEnd {
				s.LogDebugf("eq cache monitoring routine checking for a change message")
				select {
				case msg = <-c:
					haveMsg = true
				case <-time.After(time.Second * 10):
					haveMsg = false
				}
				if haveMsg {
					s.LogInfof("HEY!!!!  got a notice from the EquipmentFinder!!!!!! %+v", msg)
					if msg.ChangeType == "EquipmentChange" {
						s.RefreshCache()
					}
				}
			}
		}
		m <- "DONE"
	}(finderChannel, s.monitoringChannel)
	s.finderIF.SubscribeToChanges(finderChannel)
}
