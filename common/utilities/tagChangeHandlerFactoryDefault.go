package utilities

import (
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-common/common/core/services"
	"github.com/Spruik/libre-logging"
)

type tagChangeHandlerFactoryDefault struct {
	libreLogger.LoggingEnabler
}

func NewTagChangeHandlerFactoryDefault() *tagChangeHandlerFactoryDefault {
	var s = tagChangeHandlerFactoryDefault{}
	s.SetLoggerConfigHook("tagChangeHandlerFactory")
	return &s
}

func (s *tagChangeHandlerFactoryDefault) CreateHandlerInstance(key string, mgdEq *ports.ManagedEquipmentPort) ports.TagChangeHandlerPort {
	switch key {
	case "TagChangeHandlerPropInMemory":
		return NewTagChangeHandlerPropInMemory(mgdEq)
	case "TagChangeHandlerPackMLStatus":
		return NewTagChangeHandlerPackMLStatus(mgdEq)
	case "TagChangeHandlerPackMLAdmin":
		return NewTagChangeHandlerPackMLAdmin(mgdEq)
	case "TagChangeHandlerEventEval":
		return NewTagChangeHandlerEventEval(mgdEq,
			services.GetLibreDataStoreServiceInstance(),
			services.GetEventDefEvaluatorServiceInstance(),
			services.GetEventDefDistributorServiceInstance())
	case "TagChangeHandlerSender":
		return NewTagChangeHandlerSender(mgdEq)
	}
	return nil
}
