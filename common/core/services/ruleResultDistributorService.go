package services

import (
	"github.com/Spruik/libre-common/common/core/ports"
)

type ruleResultDistributorService struct {
	port ports.RuleResultDistributorIF
}

func NewRuleResultDistributorService(port ports.RuleResultDistributorIF) *ruleResultDistributorService {
	return &ruleResultDistributorService{
		port: port,
	}
}

var ruleResultDistributorServiceInstance *ruleResultDistributorService = nil

func SetRuleResultDistributorServiceInstance(inst *ruleResultDistributorService) {
	ruleResultDistributorServiceInstance = inst
}
func GetRuleResultDistributorServiceInstance() *ruleResultDistributorService {
	return ruleResultDistributorServiceInstance
}

func (s *ruleResultDistributorService) DistributeRuleResult(mgdEq *ports.ManagedEquipmentPort, ruleResults map[string]interface{}) error {
	return s.port.DistributeRuleResult(mgdEq, ruleResults)
}
