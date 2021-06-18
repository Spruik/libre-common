package ports

import "github.com/Spruik/libre-common/common/core/domain"

type EventDefEvaluatorPort interface {
	EvaluateEventDef(mgdEq *ManagedEquipmentPort, eventDefId string, evalContext *map[string]interface{}) (bool, *domain.EventDefinition, map[string]interface{}, error)
}
