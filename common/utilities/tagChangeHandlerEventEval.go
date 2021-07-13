package utilities

import (
	"errors"
	"fmt"
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-common/common/core/queries"
	"github.com/Spruik/libre-logging"
	"time"
)

type tagChangeHandlerEventEval struct {
	//inherit logging functions
	libreLogger.LoggingEnabler

	mgdEq                  *ports.ManagedEquipmentPort
	storeIF                ports.LibreDataStorePort
	eventDefEvalIF         ports.EventDefEvaluatorPort
	eventDefDistIF         ports.EventDefDistributorPort
	tagNameToEventDefIdMap map[string][]string
}

func NewTagChangeHandlerEventEval(mgdEq *ports.ManagedEquipmentPort, storeIF ports.LibreDataStorePort, eventDefEvalIF ports.EventDefEvaluatorPort, eventDefDistIF ports.EventDefDistributorPort) *tagChangeHandlerEventEval {
	s := tagChangeHandlerEventEval{
		mgdEq:          mgdEq,
		storeIF:        storeIF,
		eventDefEvalIF: eventDefEvalIF,
		eventDefDistIF: eventDefDistIF,
	}
	s.SetLoggerConfigHook("EVNTEVAL")
	return &s
}

func (s *tagChangeHandlerEventEval) Initialize() {
	s.tagNameToEventDefIdMap = map[string][]string{}
	//get all the event defs for this eq, it's eqc or that eqc's parents
	txn := s.storeIF.BeginTransaction(false, "evtevalinit")
	defer txn.Dispose()
	eventDefList, err := queries.GetAllEventDefsForEquipmentAndClass(txn, (*s.mgdEq).GetEquipmentId())
	if err == nil {
		//for each event def, for each trigger prop -> add to the map
		for _, evtDef := range eventDefList {
			for _, prop := range evtDef.TriggerProperties {
				evtDefIds := s.tagNameToEventDefIdMap[string(prop.Name)]
				if evtDefIds == nil {
					s.tagNameToEventDefIdMap[string(prop.Name)] = []string{}
					evtDefIds = s.tagNameToEventDefIdMap[prop.Name]
				}
				s.tagNameToEventDefIdMap[string(prop.Name)] = append(evtDefIds, evtDef.Id)
			}
		}
	} else {
		s.LogErrorf("TODO", "Error retrieving event definitions: %+v", err)
	}
}

func (s *tagChangeHandlerEventEval) HandleTagChange(tagData domain.StdMessageStruct, handlerContext *map[string]interface{}) error {
	s.LogDebug("BEGIN: tagChangeHandlerEventEval.HandleTagChange")
	var err error
	if tagData.ItemName != "" {
		_, isValid := (*s.mgdEq).GetPropertyMap()[tagData.ItemName]
		if isValid {
			entry := s.tagNameToEventDefIdMap[tagData.ItemName]
			if entry != nil {
				for _, evtDefId := range entry {
					var evalTrue bool
					var evtDef *domain.EventDefinition
					var computedFields map[string]interface{}
					evalTrue, evtDef, computedFields, err = s.eventDefEvalIF.EvaluateEventDef(s.mgdEq, evtDefId, handlerContext)
					if err == nil {
						if evalTrue {
							//update any properties that match computed fields
							for fname, fval := range computedFields {
								err = (*s.mgdEq).UpdatePropertyValue(fname, fmt.Sprintf("%s", fval))
								if err != nil {
									s.LogErrorf("Failed to add computed field %s as property of %s", fname, (*s.mgdEq).GetEquipmentName())
								}
							}

							//if evt def type is EventLog, then create an event on the equipment
							if evtDef.MessageClass == "EventLog" {
								s.addEventForEquipment((*s.mgdEq).GetEquipmentId(), evtDef, computedFields)
							}
							//distribute the event
							err = s.eventDefDistIF.DistributeEventDef((*s.mgdEq).GetEquipmentId(),(*s.mgdEq).GetEquipmentName(), evtDef, computedFields)
						}
					} else {
						s.LogErrorf("Failed in EvaluateEventDef with err=%+v", err)
					}
				}
			}

		} else {
			err = errors.New(fmt.Sprintf("Property %s is not defined for equipment %s", tagData.ItemName, tagData.OwningAsset))
		}
	}
	return err
}

func (s *tagChangeHandlerEventEval) GetAckMessage(err error) string {
	if err == nil {
		return fmt.Sprintf("\nTag change handled by evaluating Event Definitions.")
	} else {
		return fmt.Sprintf("\nFailed event evaluation while handing tag change with [%s]", err)
	}
}

func (s *tagChangeHandlerEventEval) addEventForEquipment(id string, def *domain.EventDefinition, fields map[string]interface{}) {
	_ = id
	_ = def
	_ = fields
	//Add an event instance to the target equipment using the event def info
	//TODO - need to figure out how the start/end for extended events is defined by event defs
	/*TODO - EventLog has the following to fill in:
	jobResponse:  ???
	equipment:  easy
	startDateTime:  time of the evaluation?
	endDateTime:  ????
	duration:  computed from start/end?
	reasonCode
	reasonText
	reasonValue
	reasonUoM
	comments
	*/
	evtDesc := domain.EquipmentEventDescriptor{
		Name:   def.Name,
		Time:   time.Now(), //TODO - should use some fixed time - eval time, tag change time?
		Params: fields,
	}
	_ = (*s.mgdEq).AddEvent(def.Name, evtDesc)
}
