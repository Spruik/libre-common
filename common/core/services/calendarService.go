package services

import (
	"time"

	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
)

type calendarService struct {
	dataStore     ports.CalendarPort
	publish       ports.LibreConnectorPort
	ticker        *time.Ticker
	workCalendars []domain.WorkCalendar
	eval          chan bool

	cache map[string]domain.WorkCalendarEntryType

	//inherit config functions
	libreConfig.ConfigurationEnabler

	//inherit logging functions
	libreLogger.LoggingEnabler
}

func NewCalendarService(configHook string, dataStore ports.CalendarPort, publish ports.LibreConnectorPort) *calendarService {
	var ret = calendarService{
		dataStore: dataStore,
		publish:   publish,
		ticker:    nil,
		eval:      make(chan bool),
		cache:     map[string]domain.WorkCalendarEntryType{},
	}

	ret.SetConfigCategory(configHook)
	loggerHook, cerr := ret.GetConfigItemWithDefault(domain.LOGGER_CONFIG_HOOK_TOKEN, domain.DEFAULT_LOGGER_NAME)
	if cerr != nil {
		loggerHook = domain.DEFAULT_LOGGER_NAME
	}
	ret.SetLoggerConfigHook(loggerHook)

	// TODO: Populate Cache with what ever is already in the ports.LibreConnectorPort

	return &ret
}

var calendarServiceInstance *calendarService = nil

func SetCalendarServiceInstance(inst *calendarService) {
	calendarServiceInstance = inst
}

func GetCalendarServiceInstance() *calendarService {
	return calendarServiceInstance
}

func (s *calendarService) GetAllActiveWorkCalendar() ([]domain.WorkCalendar, error) {
	return s.dataStore.GetAllActiveWorkCalendar()
}

func (s *calendarService) Start() (err error) {
	if s.ticker == nil {
		s.ticker = time.NewTicker(10 * time.Second)
		s.workCalendars, err = s.dataStore.GetAllActiveWorkCalendar()
		s.calculateCalendars()
		if err != nil {
			s.ticker.Stop()
			return err
		}
		go func() {
			for {
				select {
				case <-s.eval:
					return
				case t := <-s.ticker.C:
					s.LogDebug("Tick at ", t)
					s.calculateCalendars()
				}
			}
		}()
	} else {
		s.workCalendars, err = s.dataStore.GetAllActiveWorkCalendar()
		if err != nil {
			s.ticker.Stop()
			return err
		}
		s.ticker.Reset(0)
	}
	return nil
}

func (s *calendarService) Stop() {
	s.ticker.Stop()
	s.eval <- true
}

func (s *calendarService) calculateCalendars() {
	for _, workCalendar := range s.workCalendars {
		s.LogDebugf("processing work calendar: %s(%s)\n", workCalendar.Name, workCalendar.Id)
		if workCalendar.IsActive && len(workCalendar.Equipment) > 0 {

			// Get Current WorkCalendar EntryType
			calendarEntryType, err := workCalendar.GetCurrentEntryType()
			if err != nil {
				s.LogErrorf("%s", err)
				continue
			}

			// Inform Libre
			for _, equip := range workCalendar.Equipment {
				s.LogDebugf("\tequipment: %s(%s): is currently %s\n", equip.Name, equip.Id, calendarEntryType)

				msg := domain.StdMessageStruct{
					OwningAsset:   equip.Name,
					OwningAssetId: equip.Id,
					ItemName:      "workCalendarCategory",
					ItemNameExt:   map[string]string{},
					ItemId:        "",
					ItemValue:     string(calendarEntryType),
					ItemDataType:  "STRING",
					TagQuality:    1,
					Err:           nil,
					ChangedTime:   time.Now(),
					Category:      "WorkCalendar",
					Topic:         equip.Name + "/workCalendarCategory",
				}

				if lastState := s.cache[equip.Id]; lastState != calendarEntryType {
					msg.ItemOldValue = string(lastState)
					s.publish.SendStdMessage(msg)
					s.cache[equip.Id] = calendarEntryType
				}

			}
		}
	}
}
