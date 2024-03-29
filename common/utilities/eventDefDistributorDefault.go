package utilities

import (
	"encoding/json"
	"time"

	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/services"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
)

type eventDefDistributorDefault struct {
	libreConfig.ConfigurationEnabler
	libreLogger.LoggingEnabler
}

func NewEventDefDistributorDefault(configHook string) *eventDefDistributorDefault {
	s := eventDefDistributorDefault{}
	s.SetConfigCategory(configHook)
	loggerHook, cerr := s.GetConfigItemWithDefault(domain.LOGGER_CONFIG_HOOK_TOKEN, domain.DEFAULT_LOGGER_NAME)
	if cerr != nil {
		loggerHook = domain.DEFAULT_LOGGER_NAME
	}
	s.SetLoggerConfigHook(loggerHook)
	return &s
}

type eventStuct struct {
	Equipment string
	Event     string
	Payload   map[string]interface{}
}

func (s *eventDefDistributorDefault) DistributeEventDef(eqId, eqName string, eventDef *domain.EventDefinition, computedPayload map[string]interface{}) error {
	//by default we will just write to the log for the moment
	s.LogInfo("***********************")
	s.LogInfo("***EVENT DEF RESULT****")
	s.LogInfof("Reporting event def result for equipment %s", eqName)
	s.LogInfof("  event def %s:%s", eventDef.Id, eventDef.Name)
	s.LogInfof("     trigger: %+v", eventDef.TriggerProperties)
	s.LogInfof("     expression: %s", eventDef.TriggerExpression)
	s.LogInfof("     payload: %+v", computedPayload)
	s.LogInfo("***********************")

	//now pass it to the libre connector
	payloadData := eventStuct{
		Equipment: eqName,
		Event:     eventDef.Name,
		Payload:   computedPayload,
	}
	jsonBytes, err := json.Marshal(&payloadData)
	if err == nil {
		msg := domain.StdMessageStruct{
			OwningAsset:      eqName,
			OwningAssetId:    eqId,
			ItemName:         string(eventDef.MessageClass),
			ItemValue:        string(jsonBytes),
			ItemDataType:     "STRING",
			TagQuality:       0,
			Err:              nil,
			Category:         "EVENT",
			ChangedTimestamp: time.Now(),
		}
		err = services.GetLibreConnectorServiceInstance().SendStdMessage(msg)
	}
	return err
}
