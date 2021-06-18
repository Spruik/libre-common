package utilities

import (
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-logging"
)

type valueChangeFilterFactoryDefault struct {
	libreLogger.LoggingEnabler
}

func NewValueChangeFilterFactoryDefault() *valueChangeFilterFactoryDefault {
	var s = valueChangeFilterFactoryDefault{}
	s.SetLoggerConfigHook("valueChangeFilterFactory")
	return &s
}

func (s *valueChangeFilterFactoryDefault) CreateFilterInstance(key string, mgdEq *ports.ManagedEquipmentPort) ports.ValueChangeFilterPort {
	switch key {
	case "ValueChangeFilterDefault":
		return NewValueChangeFilterDefault()
	}
	return nil
}
