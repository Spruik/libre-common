package services

import (
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
)

type eventDefEvaluatorService struct {
	port ports.EventDefEvaluatorPort
}

func NewEventDefEvaluatorService(port ports.EventDefEvaluatorPort) *eventDefEvaluatorService {
	return &eventDefEvaluatorService{
		port: port,
	}
}

var eventDefEvaluatorServiceInstance *eventDefEvaluatorService = nil

func SetEventDefEvaluatorServiceInstance(inst *eventDefEvaluatorService) {
	eventDefEvaluatorServiceInstance = inst
}
func GetEventDefEvaluatorServiceInstance() *eventDefEvaluatorService {
	return eventDefEvaluatorServiceInstance
}

func (s *eventDefEvaluatorService) EvaluateEventDef(mgdEq *ports.ManagedEquipmentPort, eventDefId string, evalContext *map[string]interface{}) (bool, *domain.EventDefinition, map[string]interface{}, error) {
	return s.port.EvaluateEventDef(mgdEq, eventDefId, evalContext)
}
