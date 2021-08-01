package utilities

import (
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-logging"
	"sync"
)

type equipmentServiceManagerRunnerDefault struct {
	//inherit logging functions
	libreLogger.LoggingEnabler

	dataStore         ports.LibreDataStorePort
	tagChangeHandlers *[]ports.TagChangeHandlerPort
	mgdEq             *ports.ManagedEquipmentPort
}

func NewEquipmentServiceManagerRunnerDefault(loggerHook string, storeIF ports.LibreDataStorePort, tagChangeHandlers *[]ports.TagChangeHandlerPort) *equipmentServiceManagerRunnerDefault {
	s := equipmentServiceManagerRunnerDefault{
		dataStore:         storeIF,
		tagChangeHandlers: tagChangeHandlers,
	}
	s.SetLoggerConfigHook(loggerHook)
	return &s
}

func (s *equipmentServiceManagerRunnerDefault) Prepare(mgdEd *ports.ManagedEquipmentPort) {
	for _, handler := range *s.tagChangeHandlers {
		handler.Initialize()
	}
	s.mgdEq = mgdEd
}

func (s *equipmentServiceManagerRunnerDefault) Run(wg *sync.WaitGroup) {
	//called in a separate thread to handle a specific piece of equipment
	s.LogInfof("Starting equipment processing thread for equipment %s", (*s.mgdEq).GetEquipmentId())
	wg.Add(1)
	running := true
	for running {
		(*s.mgdEq).AcceptRequest(s.tagChangeHandlers)
	}
	if wg != nil {
		wg.Done()
	}
}
