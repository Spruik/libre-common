package utilities

import (
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	libreConfig "github.com/Spruik/libre-configuration"
	"github.com/Spruik/libre-logging"
)

type ruleResultDistributorDefault struct {
	libreConfig.ConfigurationEnabler
	libreLogger.LoggingEnabler
}

func NewRuleResultDistributorDefault(configHook string) *ruleResultDistributorDefault {
	s := ruleResultDistributorDefault{}
	s.SetConfigCategory(configHook)
	loggerHook, cerr := s.GetConfigItemWithDefault(domain.LOGGER_CONFIG_HOOK_TOKEN, domain.DEFAULT_LOGGER_NAME)
	if cerr != nil {
		loggerHook = domain.DEFAULT_LOGGER_NAME
	}
	s.SetLoggerConfigHook(loggerHook)
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
