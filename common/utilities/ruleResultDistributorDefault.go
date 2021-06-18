package utilities

import (
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-logging"
)

type ruleResultDistributorDefault struct {
	libreLogger.LoggingEnabler
}

func NewRuleResultDistributorDefault() *ruleResultDistributorDefault {
	s := ruleResultDistributorDefault{}
	s.SetLoggerConfigHook("EVTDISTR")
	return &s
}

func (s *ruleResultDistributorDefault) DistributeRuleResult(mgdEq *ports.ManagedEquipmentPort, ruleResults map[string]interface{}) error {
	//by default we will just write to the log for the moment
	s.LogInfo("******************")
	s.LogInfo("***RULE RESULT****")
	s.LogInfof("Reporting rule result for eq name %s", (*mgdEq).GetEquipmentName())
	s.LogInfof("  %+v", ruleResults)
	s.LogInfo("******************")
	return nil
}
