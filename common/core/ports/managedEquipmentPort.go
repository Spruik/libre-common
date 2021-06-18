package ports

import (
	"github.com/Spruik/libre-common/common/core/domain"
)

type ManagedEquipmentPort interface {
	UpdatePropertyValue(propName string, propValue interface{}) error
	AddEvent(eventName string, eventDesc domain.EquipmentEventDescriptor) error
	SetConfigLevel(level int)
	SendRequest(request domain.EquipmentServiceRequest) domain.EquipmentServiceRequest
	AcceptRequest(tagChangeHandlers *[]TagChangeHandlerPort) bool
	GetEquipmentId() string
	GetEquipmentName() string
	GetEquipmentDescription() string
	GetEquipmentLevel() string
	GetPropertyValue(propName string) interface{}
	GetPropertyMap() map[string]domain.EquipmentPropertyDescriptor
	GetEventList() *[]domain.EquipmentEventDescriptor
	GetProperty(name string) domain.EquipmentPropertyDescriptor
}

/////////////////////////////////////////////////////////////////////////////////

type ManagedEquipmentFactoryIF interface {
	GetNewInstance(eqInst domain.Equipment) ManagedEquipmentPort
}
