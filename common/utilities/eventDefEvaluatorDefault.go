package utilities

import (
	"fmt"
	"github.com/PaesslerAG/gval"
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-common/common/core/queries"
	"github.com/Spruik/libre-configuration"
	"github.com/Spruik/libre-logging"
	"strconv"
	"time"
)

type eventDefEvaluatorDefault struct {
	//inherit logging
	libreLogger.LoggingEnabler
	//inherit config
	libreConfig.ConfigurationEnabler

	dataStore ports.LibreDataStorePort
}

func NewEventDefEvaluatorDefault(configHook string, storeIF ports.LibreDataStorePort) *eventDefEvaluatorDefault {
	s := eventDefEvaluatorDefault{
		dataStore: storeIF,
	}
	s.SetConfigCategory(configHook)
	loggerHook, cerr := s.GetConfigItemWithDefault(domain.LOGGER_CONFIG_HOOK_TOKEN, domain.DEFAULT_LOGGER_NAME)
	if cerr != nil {
		loggerHook = domain.DEFAULT_LOGGER_NAME
	}
	s.SetLoggerConfigHook(loggerHook)
	return &s
}

func (s *eventDefEvaluatorDefault) EvaluateEventDef(mgdEq *ports.ManagedEquipmentPort, eventDefId string, evalContext *map[string]interface{}) (bool, *domain.EventDefinition, map[string]interface{}, error) {
	s.LogDebug("in eventDefEvaluatorDefault.EvaluateEventDef")
	var err error
	var evtDef domain.EventDefinition
	txn := s.dataStore.BeginTransaction(false, "evteval")
	defer txn.Dispose()
	evtDef, err = queries.GetEventDefinitionById(txn, eventDefId)
	if err == nil {
		var vals = make(map[string]interface{})
		for key, val := range (*mgdEq).GetPropertyMap() {
			vals[key] = val.Value
		}
		if evalContext != nil {
			for key, val := range *evalContext {
				vals[key] = val
			}
		}
		var result interface{}
		var retBool bool
		s.LogDebugf("EVALUATING [%s] with %+v", evtDef.TriggerExpression, vals)
		result, err = gval.Evaluate(evtDef.TriggerExpression, vals)
		s.LogInfof("Raw EVAL result is: %v", result)
		// a short wait to make sure that we have cached values that may be written at the same time. Example, Order Number and Material Number tags
		// might get updated in the PLC at the same time, we might use Order Number in the trigger and material number in the payload, we want to
		// make sure that we have cached the updated material number before we build the payload.
		delay, _ := s.GetConfigItem("EventPayloadDelayMilliseconds")
		delayDuration,_:= time.ParseDuration(delay)
		time.Sleep(delayDuration)
		for key, val := range (*mgdEq).GetPropertyMap() {
			vals[key] = val.Value
		}
		if evalContext != nil {
			for key, val := range *evalContext {
				vals[key] = val
			}
		}
		if err == nil {
			switch v := result.(type) {
			default:
				s.LogErrorf("unexpected type %T", v)
			case bool:
				boolStr := fmt.Sprintf("%t", result)
				retBool, err = strconv.ParseBool(boolStr) //HACKALERT - how to do this properly?
				s.LogInfof("parsed EVAL result is: %v", retBool)
				if err == nil {
					if retBool {
						s.LogInfof("parsed EVAL result tested TRUE")
						//result of evaluation is "TRUE"
						//need to build up the map of payload fields to return - both from property values and field expressions
						fieldMap := map[string]interface{}{}
						var fieldVal interface{}
						var fieldErr error
						for _, field := range evtDef.PayloadFields {
							fieldVal, fieldErr = gval.Evaluate(field.Expression, vals)
							if fieldErr == nil {
								//add to our field map
								fieldMap[field.Name] = fieldVal
							} else {
								s.LogErrorf("Failed in gval Evaluate of payload field %s with err=%+v", field.Name, fieldErr)
								break
							}
						}
						for _, evtPayProp := range evtDef.PayloadProperties {
							fieldMap[evtPayProp.Name] = (*mgdEq).GetPropertyValue(evtPayProp.Name)
						}
						return retBool, &evtDef, fieldMap, fieldErr
					} else {
						return retBool, &evtDef, nil, nil
					}
				} else {
					s.LogErrorf("Failed in parse boolean value '%s' with err=%+v", boolStr, err)
				}
			}
		} else {
			s.LogErrorf("Failed in gval Evaluate with err=%+v", err)
		}
	} else {
		s.LogErrorf("Failed in GetEventDefinitionById with err=%+v", err)
	}
	return false, nil, nil, err
}
