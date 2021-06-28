package utilities

import (
	"fmt"
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-logging"
)

type tagChangeHandlerPackMLAdmin struct {
	//inherit logging functions
	libreLogger.LoggingEnabler

	mgdEq *ports.ManagedEquipmentPort
}

func NewTagChangeHandlerPackMLAdmin(mgdEq *ports.ManagedEquipmentPort) *tagChangeHandlerPackMLAdmin {
	s := tagChangeHandlerPackMLAdmin{
		mgdEq: mgdEq,
	}
	s.SetLoggerConfigHook("TAGCHNG")
	return &s
}

func (s *tagChangeHandlerPackMLAdmin) Initialize() {
	//TODO - implementation TBD
}

func (s *tagChangeHandlerPackMLAdmin) HandleTagChange(tagData domain.StdMessageStruct, handlerContext *map[string]interface{}) error {
	s.LogDebug("BEGIN: tagChangeHandlerPackMLAdmin.HandleTagChange")
	_, pcexists := tagData.ItemNameExt["PRODCONSCNT"] //TODO - get this token from configuration
	_, ppexists := tagData.ItemNameExt["PRODPROCCNT"] //TODO - get this token from configuration
	_, pdexists := tagData.ItemNameExt["PRODDFCTCNT"] //TODO - get this token from configuration
	//if these tokens are present, then we are dealing with a PackML admin message
	if pcexists || ppexists || pdexists {
		//TODO - implementation TDB
		s.LogInfof("Handling a PackML Admin message for %s with: %+v", (*s.mgdEq).GetEquipmentName(), tagData.ItemNameExt)
	}
	return nil
}

func (s *tagChangeHandlerPackMLAdmin) GetAckMessage(err error) string {
	if err == nil {
		return fmt.Sprintf("\nTag change handled by as a PackML status message.")
	} else {
		return fmt.Sprintf("\nFailed PackML status message processing while handling tag change with error [%s]", err)
	}
}
