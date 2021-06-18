package ports

import (
	"github.com/Spruik/libre-common/common/core/domain"
)

//The TagChangeHandlerPort interface defines the functions to be provided for handling a tag change message within
//  the EquipmentTagManagerService
type TagChangeHandlerPort interface {
	Initialize()
	HandleTagChange(tagData domain.StdMessageStruct, handlerContext *map[string]interface{}) error
	GetAckMessage(err error) string
}

//////////////////////////////////

type TagChangeHandlerFactoryPort interface {
	CreateHandlerInstance(key string, mgdEq *ManagedEquipmentPort) TagChangeHandlerPort
}
