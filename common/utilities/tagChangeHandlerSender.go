package utilities

import (
	"fmt"

	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-common/common/core/services"
	libreLogger "github.com/Spruik/libre-logging"
)

type tagChangeHandlerSender struct {
	//inherit logging functions
	libreLogger.LoggingEnabler

	mgdEq *ports.ManagedEquipmentPort
}

func NewTagChangeHandlerSender(mgdEq *ports.ManagedEquipmentPort) *tagChangeHandlerSender {
	s := tagChangeHandlerSender{
		mgdEq: mgdEq,
	}
	s.SetLoggerConfigHook("TAGCHNG")
	return &s
}

func (s *tagChangeHandlerSender) Initialize() {

}

func (s *tagChangeHandlerSender) HandleTagChange(tagData domain.StdMessageStruct, handlerContext *map[string]interface{}) error {
	s.LogDebug("BEGIN: tagChangeHandlerSender.HandleTagChange")
	return services.GetLibreConnectorServiceInstance().SendStdMessage(tagData)
}

func (s *tagChangeHandlerSender) GetAckMessage(err error) string {
	if err == nil {
		return "\nTag change handled by sending tag data to the center."
	} else {
		return fmt.Sprintf("\nFailed to send tag change while handing tag change with error [%s]", err)
	}
}
