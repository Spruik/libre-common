package utilities

import (
	"fmt"
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-logging"
)

type tagChangeHandlerPackMLStatus struct {
	//inherit logging functions
	libreLogger.LoggingEnabler

	mgdEq *ports.ManagedEquipmentPort
}

func NewTagChangeHandlerPackMLStatus(mgdEq *ports.ManagedEquipmentPort) *tagChangeHandlerPackMLStatus {
	s := tagChangeHandlerPackMLStatus{
		mgdEq: mgdEq,
	}
	s.SetLoggerConfigHook("TAGCHNG")
	return &s
}

func (s *tagChangeHandlerPackMLStatus) Initialize() {
	//TODO - implementation TBD
}

func (s *tagChangeHandlerPackMLStatus) HandleTagChange(tagData domain.StdMessageStruct, handlerContext *map[string]interface{}) error {
	s.LogDebug("BEGIN: tagChangeHandlerPackMLStatus.HandleTagChange")
	_, paramexists := tagData.ItemNameExt["PARAMNUM"]  //TODO - get this token from configuration
	_, prodexists := tagData.ItemNameExt["PRODUCTNUM"] //TODO - get this token from configuration
	//if these tokens exist, then we are dealing with a PackML status message
	if paramexists || prodexists {
		//TODO - implementation TDB
		s.LogInfof("Handling a PackML Status message for %s with: %+v", (*s.mgdEq).GetEquipmentName(), tagData.ItemNameExt)
	}
	return nil
}

func (s *tagChangeHandlerPackMLStatus) GetAckMessage(err error) string {
	if err == nil {
		return fmt.Sprintf("\nTag change handled by as a PackML status message.")
	} else {
		return fmt.Sprintf("\nFailed PackML status message processing while handling tag change with error [%s]", err)
	}
}
