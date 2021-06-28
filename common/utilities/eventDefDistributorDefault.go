package utilities

import (
	"encoding/json"
	"fmt"
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/services"
	"github.com/Spruik/libre-logging"
)

type eventDefDistributorDefault struct {
	libreLogger.LoggingEnabler
}

func NewEventDefDistributorDefault() *eventDefDistributorDefault {
	s := eventDefDistributorDefault{}
	s.SetLoggerConfigHook("EVTDISTR")
	return &s
}

type eventStuct struct {
	Equipment string
	Event     string
	Payload   map[string]interface{}
}

func (s *eventDefDistributorDefault) DistributeEventDef(eqName string, eventDef *domain.EventDefinition, computedPayload map[string]interface{}) error {
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
			OwningAsset:  eqName,
			ItemName:     fmt.Sprintf("%s", eventDef.MessageClass),
			ItemValue:    string(jsonBytes),
			ItemDataType: "STRING",
			TagQuality:   0,
			Err:          nil,
			Category:     "EVENT",
		}
		err = services.GetLibreConnectorServiceInstance().SendStdMessage(msg)
	}
	return err
}
