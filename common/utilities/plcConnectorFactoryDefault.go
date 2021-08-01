package utilities

import (
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-common/common/drivers"
)

type plcConnectorFactoryDefault struct {
}

func NewPlcConnectorFactoryDefault() *plcConnectorFactoryDefault {
	var s = plcConnectorFactoryDefault{}
	return &s
}

func (s *plcConnectorFactoryDefault) CreatePlcConnectorInstance(key string, configHook string) ports.PlcConnectorPort {
	switch key {
	case "PlcConnectorMQTT":
		return drivers.NewPlcConnectorMQTT(configHook)
	case "PlcConnectorMQTTv3":
		return drivers.NewPlcConnectorMQTTv3(configHook)
	case "PlcConnectorOPCUA":
		return drivers.NewPlcConnectorOPCUA(configHook)
	}
	return nil
}
