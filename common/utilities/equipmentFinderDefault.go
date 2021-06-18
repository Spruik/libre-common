package utilities

import (
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-common/common/core/queries"
	"github.com/Spruik/libre-configuration"
	"github.com/Spruik/libre-logging"
	"strings"
)

type equipmentFinderDefault struct {
	//inherit logging functions
	libreLogger.LoggingEnabler

	//inherit config functions
	libreConfig.ConfigurationEnabler

	dataStore ports.LibreDataStorePort
}

func NewEquipmentFinderDefault(storeIF ports.LibreDataStorePort) *equipmentFinderDefault {
	s := equipmentFinderDefault{
		dataStore: storeIF,
	}
	s.SetLoggerConfigHook("EQFINDR")
	s.SetConfigCategory("equipmentFinder")
	return &s
}

func (s *equipmentFinderDefault) FindEquipment() ([]domain.Equipment, error) {
	var ret = make([]domain.Equipment, 0, 0)
	eqSet := make(map[string]domain.Equipment)
	var err error
	txn := s.dataStore.BeginTransaction(false, "findeq")
	defer txn.Dispose()
	//get the equipment type list from config
	levelList, _ := s.GetConfigItemWithDefault("ACTIVE_EQ_LEVELS", "")
	if levelList != "" {
		levelSlice := strings.Split(levelList, ",")
		for _, level := range levelSlice {
			//query for the matching equipment
			var eqs []domain.Equipment
			eqs, err = queries.GetActiveEquipmentByLevel(txn, level)
			if err == nil {
				for _, eq := range eqs {
					eqSet[eq.Id] = eq
				}
			}
		}
	}
	incList, _ := s.GetConfigItemWithDefault("INCLUDE_EQUIPMENT", "")
	if incList != "" {
		incSlice := strings.Split(incList, ",")
		for _, eqName := range incSlice {
			//query for the matching equipment
			var eq domain.Equipment
			eq, err = queries.GetEquipmentByName(txn, eqName)
			if err == nil {
				eqSet[eq.Id] = eq
			}
		}
	}
	excList, _ := s.GetConfigItemWithDefault("EXCLUDE_EQUIPMENT", "")
	if excList != "" {
		excSlice := strings.Split(excList, ",")
		for _, eqName := range excSlice {
			//query for the matching equipment
			var eq domain.Equipment
			eq, err = queries.GetEquipmentByName(txn, eqName)
			if err == nil {
				eqSet[eq.Id] = domain.Equipment{}
			}
		}
	}
	for _, eq := range eqSet {
		if eq.Id != "" {
			ret = append(ret, eq)
		}
	}
	return ret, err
}
