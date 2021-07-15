package utilities

import (
	"encoding/json"
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-common/common/core/queries"
	"github.com/Spruik/libre-configuration"
	"github.com/Spruik/libre-logging"
	"strings"
	"time"
)

type equipmentFinderDefault struct {
	//inherit logging functions
	libreLogger.LoggingEnabler

	//inherit config functions
	libreConfig.ConfigurationEnabler

	dataStore        ports.LibreDataStorePort
	subPort          ports.LibreDataStoreSubscriptionPort
	daSubChan        chan []byte
	monitorAdminChan chan string
	monitoring       bool
}

func NewEquipmentFinderDefault(storeIF ports.LibreDataStorePort) *equipmentFinderDefault {
	s := equipmentFinderDefault{
		dataStore:  storeIF,
		monitoring: false,
	}
	s.SetConfigCategory("equipmentFinder")
	s.SetLoggerConfigHook("EQFINDR")
	return &s
}

func (s *equipmentFinderDefault) FindEquipment() ([]domain.Equipment, error) {
	txn := s.dataStore.BeginTransaction(false, "findeq")
	defer txn.Dispose()

	eqElLevels, includeIds, excludeIds, err := s.getQueryInput(txn)

	eqs, err := queries.GetActiveEquipmentByLevelListWithIncExc(txn, eqElLevels, includeIds, excludeIds)
	return eqs, err
}

func (s *equipmentFinderDefault) UnsubscribeToChanges(notificationChannel chan ports.EquipmentFinderChangeNotice) {
	s.LogDebug("Equipment Finder UnsubscribeToChanges begins")
	s.monitoring = false
	s.subPort.StopGettingSubscriptionNotifications()
	s.daSubChan = nil
	s.monitorAdminChan <- "END"
	select {
	case resp := <-s.monitorAdminChan:
		if resp == "DONE" {
			s.LogInfo("equipment finder monitor routine ended successfully")
		} else {
			s.LogErrorf("equipment finder monitor routine got bad value for ack at shutdown: %s", resp)
		}
	case <-time.After(5 * time.Second):
		s.LogErrorf("equipment finder monitor routine timed out at shutdown")
	}
	s.LogDebug("Equipment Finder UnsubscribeToChanges ends")
}

func (s *equipmentFinderDefault) SubscribeToChanges(notificationChannel chan ports.EquipmentFinderChangeNotice) {
	txn := s.dataStore.BeginTransaction(false, "findeq4s")
	defer txn.Dispose()
	eqElLevels, includeIds, excludeIds, err := s.getQueryInput(txn)
	if err == nil {
		vars := map[string]interface{}{"levels": eqElLevels, "includeIds": includeIds, "excludeIds": excludeIds}
		var q struct {
			QueryEquipment []domain.Equipment `graphql:"queryEquipment (filter:{isActive: true, and: {equipmentLevel: {in: $levels}, or: {id:$includeIds}, and: {not:{id:$excludeIds}}}}) "`
		}
		s.subPort = s.dataStore.GetSubscription(&q, vars)
		s.daSubChan = make(chan []byte)
		s.monitorAdminChan = make(chan string)
		s.monitoring = true
		go func(c chan []byte, n chan ports.EquipmentFinderChangeNotice, a chan string) {
			a <- "RUNNING"
			var shouldEnd = false
			var haveDBMsg bool
			var msg []byte
			var haveAdminMsg bool
			var adminMsg string
			for s.monitoring && !shouldEnd {
				select {
				case msg = <-c:
					haveDBMsg = true
					haveAdminMsg = false
				case adminMsg = <-a:
					haveDBMsg = false
					haveAdminMsg = true
				case <-time.After(time.Second * 10):
					haveDBMsg = false
					haveAdminMsg = false
				}
				if haveAdminMsg {
					s.LogDebug("Hey!!!!!! got an admin message!!!!!!!!!!!")
					if adminMsg == "END" {
						s.LogDebug("Got and END admin message")
						shouldEnd = true
					}
				}
				if haveDBMsg {
					s.LogDebugf("Hey!!!!!! got a notice from the database for an eq change!!!!!!!!!!!")
					var q struct {
						QueryEquipment []domain.Equipment `json:"queryEquipment"`
					}
					err := json.Unmarshal(msg, &q)
					var ret = ports.EquipmentFinderChangeNotice{
						ChangeType: "EquipmentChange",
						Equipment:  q.QueryEquipment,
					}
					if err == nil {
						n <- ret
					} else {
						s.LogErrorf("Failed in Unmarshal of eq change message: %s", err)
					}
				} else {
					s.LogDebugf("waiting for change notice in equipment finder")
				}
			}
			a <- "DONE"
		}(s.daSubChan, notificationChannel, s.monitorAdminChan)
		resp := ""
		select {
		case resp = <-s.monitorAdminChan:
			if resp == "RUNNING" {
				s.LogInfo("monitor routine for equipment finder is running")
			} else {
				s.LogErrorf("monitor routine for equipment finder gave bad ack at start: %s", resp)
			}
		case <-time.After(2 * time.Second):
			s.LogError("monitor routine for equipment finder timed out at start")
		}
		s.subPort.GetSubscriptionNotifications(s.daSubChan)
	} else {
		panic(err)
	}
}

func (s *equipmentFinderDefault) getQueryInput(txn ports.LibreDataStoreTransactionPort) ([]domain.EquipmentElementLevel, []string, []string, error) {
	var err error
	//get the equipment type list from config
	eqElLevels := make([]domain.EquipmentElementLevel, 0, 0)
	var eqElLevel domain.EquipmentElementLevel
	levelList, _ := s.GetConfigItemWithDefault("ACTIVE_EQ_LEVELS", "")
	if levelList != "" {
		levelSlice := strings.Split(levelList, ",")
		for _, level := range levelSlice {
			eqElLevel = domain.EquipmentElementLevel(level)
			eqElLevels = append(eqElLevels, eqElLevel)
		}
	}

	//get the includes from config
	incList, _ := s.GetConfigItemWithDefault("INCLUDE_EQUIPMENT", "")
	includeIds := make([]string, 0, 0)
	if incList != "" {
		incSlice := strings.Split(incList, ",")
		for _, eqName := range incSlice {
			//query for the matching equipment
			var eq domain.Equipment
			eq, err = queries.GetEquipmentByName(txn, eqName)
			if err == nil {
				includeIds = append(includeIds, eq.Id)
			}
		}
	}

	//get the excludes from config
	excList, _ := s.GetConfigItemWithDefault("EXCLUDE_EQUIPMENT", "")
	excludeIds := make([]string, 0, 0)
	if excList != "" {
		excSlice := strings.Split(excList, ",")
		for _, eqName := range excSlice {
			//query for the matching equipment
			var eq domain.Equipment
			eq, err = queries.GetEquipmentByName(txn, eqName)
			if err == nil {
				excludeIds = append(excludeIds, eq.Id)
			}
		}
	}
	return eqElLevels, includeIds, excludeIds, err
}
