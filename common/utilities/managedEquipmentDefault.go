package utilities

import (
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-common/common/core/queries"
	"github.com/Spruik/libre-configuration"
	"github.com/Spruik/libre-logging"
	"sync"
	"time"
)

type managedEquipmentDefault struct {
	//inherit logging functions
	libreLogger.LoggingEnabler

	//inherit config functions
	libreConfig.ConfigurationEnabler
	mu sync.Mutex
	EquipInst      domain.Equipment
	ConfigLevel    int
	RequestChannel chan domain.EquipmentServiceRequest
	props          map[string]domain.EquipmentPropertyDescriptor
	events         []domain.EquipmentEventDescriptor
}

func NewManagedEquipmentDefault(configHook string, eqInst domain.Equipment, dataStore ports.LibreDataStorePort) *managedEquipmentDefault {
	s := managedEquipmentDefault{
		EquipInst:      eqInst,
		ConfigLevel:    0,
		RequestChannel: make(chan domain.EquipmentServiceRequest),
		props:          map[string]domain.EquipmentPropertyDescriptor{},
		events:         make([]domain.EquipmentEventDescriptor, 0, 0),
	}
	s.SetConfigCategory(configHook)
	loggerHook, cerr := s.GetConfigItemWithDefault(domain.LOGGER_CONFIG_HOOK_TOKEN, domain.DEFAULT_LOGGER_NAME)
	if cerr != nil {
		loggerHook = domain.DEFAULT_LOGGER_NAME
	}
	s.SetLoggerConfigHook(loggerHook)

	txn := dataStore.BeginTransaction(false, "getProp"+s.GetEquipmentName())
	defer txn.Dispose()
	proplist, err := getAllPropertiesForEquipment(txn, eqInst.Id)
	if err == nil {
		var props = map[string]domain.EquipmentPropertyDescriptor{}
		var val interface{} = nil
		for name, prop := range proplist {
			val, err = domain.ConvertPropertyValueStringToTypedValue(prop.DataType, prop.Value)
			if err != nil {
				s.LogErrorf("Failed data format conversion for property %s with value string %s.  Error=%s", prop.Name, prop.Value, err)
			}
			clsPropId := ""
			if prop.EquipmentClass.Id != "" {
				clsPropId = prop.Id
			}
			props[name] = domain.EquipmentPropertyDescriptor{
				Name:             name,
				DataType:         prop.DataType,
				Value:            val,
				ClassPropertyId:  clsPropId,
				EquipmentClassId: prop.EquipmentClass.Id,
				LastUpdate:       time.Time{},
			}
		}
		s.props = props
	}
	return &s
}

func getAllPropertiesForEquipment(txn ports.LibreDataStoreTransactionPort, eqId string) (map[string]domain.Property, error) {
	//need to look for properties attached to the Equipment and to it's equipment class (and equipment class parents)
	var fullPropertyList = map[string]domain.Property{}
	//first check equipment
	var eqInst domain.Equipment
	var err error = nil
	eqInst, err = queries.GetEquipmentById(txn, eqId)
	if err == nil {
		for _, p := range eqInst.Properties {
			fullPropertyList[p.Name] = p
		}

		//now work up through the equipment classes
		currEqcId := eqInst.EquipmentClass.Id
		var eqcInst domain.EquipmentClassPropertiesAndParent
		for currEqcId != "" {
			eqcInst, err = queries.GetEquipmentClassPropertiesAndParentById(txn, currEqcId)
			if err == nil {
				for _, p := range eqcInst.Properties {
					fullPropertyList[p.Name] = p
				}
				currEqcId = eqcInst.Parent.Id
			} else {
				currEqcId = ""
			}
		}
	}
	return fullPropertyList, err
}

func (s *managedEquipmentDefault) UpdatePropertyValue(propName string, propValue interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	pd, exists := s.props[propName]
	if exists {
		val, err := domain.ConvertPropertyValueStringToTypedValue(pd.DataType, propValue)
		if err == nil {
			pd.Value = val
			pd.LastUpdate = time.Now()
			s.props[propName] = pd
		}
		s.LogInfof("Property update: %s %+v (%T) @ %s", propName, pd.Value, pd.Value, pd.LastUpdate)
	} else {
		//choosing to ignore updates for unknown properties (event evaluator attempts to update payload references)
		s.LogDebugf("ignoring update request for unknown property: %s=%s", propName, propValue)
	}
	return nil
}

func (s *managedEquipmentDefault) AddEvent(eventName string, eventDesc domain.EquipmentEventDescriptor) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = eventName
	//TODO - figure out event strucuture (start/end?)
	s.events = append(s.events, eventDesc)
	return nil
}

func (s *managedEquipmentDefault) GetPropertyValue(propName string) interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.props[propName].Value
}
func (s *managedEquipmentDefault) GetProperty(name string) domain.EquipmentPropertyDescriptor {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.props[name]
}

func (s *managedEquipmentDefault) SetConfigLevel(level int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ConfigLevel = level
}

func (s *managedEquipmentDefault) GetConfigLevel() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ConfigLevel
}

func (s *managedEquipmentDefault) GetEquipmentId() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.EquipInst.Id
}

func (s *managedEquipmentDefault) GetEquipmentName() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.EquipInst.Name
}
func (s *managedEquipmentDefault) GetEquipmentDescription() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.EquipInst.Description
}
func (s *managedEquipmentDefault) GetEquipmentLevel() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return string(s.EquipInst.EquipmentLevel)
}

func (s *managedEquipmentDefault) SendRequest(request domain.EquipmentServiceRequest) domain.EquipmentServiceRequest {
	s.RequestChannel <- request
	ack := <-s.RequestChannel
	return ack
}

func (s *managedEquipmentDefault) AcceptRequest(tagChangeHandlers *[]ports.TagChangeHandlerPort) bool {
	rqst := <-s.RequestChannel
	s.LogDebugf("Managed equipment %s received request through channel: %+v", s.EquipInst.Name, rqst)
	switch rqst.ServiceType {
	case domain.SVCRQST_TAGDATA:

		var ackMsg string
		handlerContext := make(map[string]interface{})
		rqst.TagInfo.OwningAssetId = s.EquipInst.Id
		for _, handler := range *tagChangeHandlers {
			err := handler.HandleTagChange(rqst.TagInfo, &handlerContext)
			if err != nil {
				s.LogErrorf("Failed to update Equipment Property from tag %+v with error: %s", rqst.TagInfo, err)
			}
			ackMsg += handler.GetAckMessage(err)
		}
		s.RequestChannel <- domain.EquipmentServiceRequest{
			ServiceType: domain.SVCRQST_TAGDATA_ACK,
			Time:        time.Now(),
			Message:     ackMsg,
			TagInfo:     domain.StdMessageStruct{},
		}

	case domain.SVCRQST_SHUTDOWN:

		s.LogInfof("Processing thread for equipment with name=%s is shutting down", s.EquipInst.Name)
		s.RequestChannel <- domain.EquipmentServiceRequest{
			ServiceType: domain.SVCRQST_SHUTDOWN_ACK,
			Time:        time.Now(),
			Message:     "Shutdown request acknowledged",
			TagInfo:     domain.StdMessageStruct{},
		}
		return false
	}
	return true
}

func (s *managedEquipmentDefault) GetPropertyMap() map[string]domain.EquipmentPropertyDescriptor {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.props
}

func (s *managedEquipmentDefault) GetEventList() *[]domain.EquipmentEventDescriptor {
	s.mu.Lock()
	defer s.mu.Unlock()
	return &s.events
}

///////////////////////////////////////////////////////////////////////

type managedEquipmentFactoryDefault struct {
	dataStore          ports.LibreDataStorePort
	instanceConfigHook string
}

func NewManagedEquipmentFactoryDefault(dataStore ports.LibreDataStorePort, instanceConfigHook string) *managedEquipmentFactoryDefault {
	return &managedEquipmentFactoryDefault{
		dataStore:          dataStore,
		instanceConfigHook: instanceConfigHook,
	}
}

func (s *managedEquipmentFactoryDefault) GetNewInstance(eqInst domain.Equipment) ports.ManagedEquipmentPort {
	return NewManagedEquipmentDefault(s.instanceConfigHook, eqInst, s.dataStore)
}
