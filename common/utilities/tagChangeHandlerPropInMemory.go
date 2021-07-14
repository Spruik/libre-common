package utilities

import (
	"fmt"
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-logging"
)

type tagChangeHandlerPropInMemory struct {
	//inherit logging functions
	libreLogger.LoggingEnabler

	mgdEq *ports.ManagedEquipmentPort
}

func NewTagChangeHandlerPropInMemory(mgdEq *ports.ManagedEquipmentPort) *tagChangeHandlerPropInMemory {
	s := tagChangeHandlerPropInMemory{
		mgdEq: mgdEq,
	}
	s.SetLoggerConfigHook("TAGCHNG")
	return &s
}

func (s *tagChangeHandlerPropInMemory) Initialize() {

}

func (s *tagChangeHandlerPropInMemory) HandleTagChange(tagData domain.StdMessageStruct, handlerContext *map[string]interface{}) error {
	s.LogDebug("BEGIN: tagChangeHandlerPropInMemory.HandleTagChange")
	if tagData.ItemName != "" {
		oldVal := (*s.mgdEq).GetPropertyValue(tagData.ItemName)
		if oldVal != nil {
			tagData.ItemOldValue = fmt.Sprintf("%s", oldVal)
		}
		return (*s.mgdEq).UpdatePropertyValue(tagData.ItemName, tagData.ItemValue)
	}
	return nil
}

func (s *tagChangeHandlerPropInMemory) GetAckMessage(err error) string {
	if err == nil {
		return fmt.Sprintf("\nTag change handled by updating matching equipment property.")
	} else {
		return fmt.Sprintf("\nFailed equipment property update while handing tag change with error [%s]", err)
	}
}
