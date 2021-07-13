package services

import (
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
)

type eventDefDistributorService struct {
	port ports.EventDefDistributorPort
}

func NewEventDefDistributorService(port ports.EventDefDistributorPort) *eventDefDistributorService {
	return &eventDefDistributorService{
		port: port,
	}
}

var eventDefDistributorServiceInstance *eventDefDistributorService = nil

func SetEventDefDistributorServiceInstance(inst *eventDefDistributorService) {
	eventDefDistributorServiceInstance = inst
}
func GetEventDefDistributorServiceInstance() *eventDefDistributorService {
	return eventDefDistributorServiceInstance
}

func (s *eventDefDistributorService) DistributeEventDef(eqId, eqName string, eventDef *domain.EventDefinition, computedPayload map[string]interface{}) error {
	return s.port.DistributeEventDef(eqId, eqName, eventDef, computedPayload)
}
