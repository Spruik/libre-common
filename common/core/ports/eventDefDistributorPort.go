package ports

import "github.com/Spruik/libre-common/common/core/domain"

type EventDefDistributorPort interface {
	DistributeEventDef(eqName string, eventDef *domain.EventDefinition, computedPayload map[string]interface{}) error
}
