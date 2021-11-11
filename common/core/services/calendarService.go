package services

import (
	"strings"
	"time"

	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
)

const WorkCalendarCategory = "workCalendarCategory"
const WorkCalendarEntry = "workCalendarEntry"

type calendarService struct {
	dataStore      ports.CalendarPort
	publish        ports.EdgeConnectorPort
	ticker         *time.Ticker
	workCalendars  []domain.WorkCalendar
	eval           chan bool
	tickerDuration time.Duration

	cacheType    map[string]domain.WorkCalendarEntryType
	cacheEntries map[string]string

	//inherit config functions
	libreConfig.ConfigurationEnabler

	//inherit logging functions
	libreLogger.LoggingEnabler

	// Hydrate Cache Subscriptions
	hydrateUpdate chan domain.StdMessageStruct
	updates       int
}

// NewCalendarService bootstraps and creates a new Calendar Service
func NewCalendarService(configHook string, dataStore ports.CalendarPort, publish ports.EdgeConnectorPort) *calendarService {
	var ret = calendarService{
		dataStore:      dataStore,
		publish:        publish,
		ticker:         nil,
		eval:           make(chan bool),
		cacheType:      map[string]domain.WorkCalendarEntryType{},
		cacheEntries:   map[string]string{},
		tickerDuration: time.Second * 60,
	}

	ret.SetConfigCategory(configHook)
	loggerHook, cerr := ret.GetConfigItemWithDefault(domain.LOGGER_CONFIG_HOOK_TOKEN, domain.DEFAULT_LOGGER_NAME)
	if cerr != nil {
		loggerHook = domain.DEFAULT_LOGGER_NAME
	}
	ret.SetLoggerConfigHook(loggerHook)

	tickerDuration, err := ret.GetConfigItemWithDefault("tickDuration", "60s")
	if err != nil {
		ret.LogWarnf("failed to parse calendarService tickSpeed from configuration into duration: %s", tickerDuration)
	}

	dur, err := time.ParseDuration(tickerDuration)
	if err != nil {
		ret.LogWarnf("failed to parse calendarService tickSpeed %s into duration; using default %s", tickerDuration, ret.tickerDuration)
	} else {
		ret.tickerDuration = dur
	}

	return &ret
}

var calendarServiceInstance *calendarService = nil

// SetCalendarServiceInstance sets the current calendar service for this scope
func SetCalendarServiceInstance(inst *calendarService) {
	calendarServiceInstance = inst
}

// GetCalendarServiceInstance gets the current calendar service for this scope
func GetCalendarServiceInstance() *calendarService {
	return calendarServiceInstance
}

func (s *calendarService) SetTickSpeed(dur time.Duration) {
	s.tickerDuration = dur
}

func (s *calendarService) GetAllActiveWorkCalendar() ([]domain.WorkCalendar, error) {
	return s.dataStore.GetAllActiveWorkCalendar()
}

func (s *calendarService) hydrateCache() error {

	// Initialize variables to track hydration
	s.hydrateUpdate = make(chan domain.StdMessageStruct, 100)
	s.updates = 0

	// Get All the Work Calendars
	workCalendars, err := s.dataStore.GetAllActiveWorkCalendar()
	if err != nil {
		s.LogErrorf("calendarService failed to hydrate cache, expected no error getting AllActiveWorkCalendar; got %s", err)
		return err
	}

	// Find the Unique Equipment
	equipment := make([]domain.Equipment, 0)
	for _, workCalendar := range workCalendars {
		equipment = append(equipment, workCalendar.Equipment...)
	}
	deduplicateEquipment := domain.DeduplicateEquipment(equipment)

	// Early return if no equipment defined
	if len(deduplicateEquipment) == 0 {
		close(s.hydrateUpdate)
		return nil
	}

	// Subscribe to Tag Value Change
	for _, equipment := range deduplicateEquipment {
		changeFilter := map[string]interface{}{
			"EQ":     equipment.Name,
			"Client": equipment.Name,
		}
		s.publish.ListenForEdgeTagChanges(s.hydrateUpdate, changeFilter)
	}

	// Wait for All Responses or timeout
	expectedMessageCount := len(deduplicateEquipment) * 2
	s.listenForHyrdateResponses(expectedMessageCount)

	if s.updates != expectedMessageCount {
		s.LogWarnf("Failed to hydrate cache for all equipment")
	}

	// Unsubscribe from equipment
	for _, equipment := range deduplicateEquipment {
		err := s.publish.StopListeningForTagChanges(equipment.Name)
		if err != nil {
			s.LogDebugf("calendarService failed to unsubscribe to cache hydrate for %s, expected no error; got %s", equipment.Name, err)
		}
	}

	s.updates = 0
	close(s.hydrateUpdate)
	return nil
}

func (s *calendarService) listenForHyrdateResponses(expectedMessageCount int) {
	for {
		select {
		case update := <-s.hydrateUpdate:
			if update.ItemName == WorkCalendarEntry {
				s.cacheEntries[update.OwningAssetId] = update.ItemValue.(string)
				s.updates++
			}

			if update.ItemName == WorkCalendarCategory && update.ItemDataType == domain.DataTypeString {
				valueAsString := update.ItemValue.(string)
				switch valueAsString {
				case string(domain.PlannedBusyTime):
					s.cacheType[update.OwningAssetId] = domain.PlannedBusyTime
				case string(domain.PlannedDowntime):
					s.cacheType[update.OwningAssetId] = domain.PlannedDowntime
				case string(domain.PlannedShutdown):
					s.cacheType[update.OwningAssetId] = domain.PlannedShutdown
				}
				s.updates++
			}

			// If we have hit all of them we can exit early
			if s.updates == expectedMessageCount {
				return
			}
		case <-time.After(3 * time.Second):
			return
		}
	}
}

func (s *calendarService) Start() (err error) {
	s.hydrateCache()
	if s.ticker == nil {
		s.ticker = time.NewTicker(s.tickerDuration)
		s.workCalendars, err = s.dataStore.GetAllActiveWorkCalendar()
		s.calculateCalendars()
		if err != nil {
			s.LogErrorf("calendarService tick failed; got %s", err)
			return err
		}
		s.LogInfo("calendar service started")
		go func() {
		outside:
			for {
				select {
				case <-s.eval:
					break outside
				case t := <-s.ticker.C:
					s.LogDebugf("Tick at %s\n", t)
					s.workCalendars, err = s.dataStore.GetAllActiveWorkCalendar()
					if err != nil {
						s.LogErrorf("calendarService failed to get all active work calendars; got %s", err)
					}
					s.calculateCalendars()
				}
			}
			s.ticker = nil // Safe to nil this as we should no longer be listening to events
			s.LogInfo("calendar service stopped")
		}()
	} else {
		s.workCalendars, err = s.dataStore.GetAllActiveWorkCalendar()
		if err != nil {
			s.LogErrorf("calendarService tick failed; got %s", err)
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
		s.LogDebugf("processing work calendar: %s(%s)\n", workCalendar.Name, workCalendar.ID)
		if workCalendar.IsActive && len(workCalendar.Equipment) > 0 {

			// Get Current WorkCalendar EntryType
			calendarEntryType, names, err := workCalendar.GetCurrentEntryTypeAndNames()
			if err != nil {
				s.LogErrorf("%s", err)
				continue
			}

			calendarEntries := strings.Join(names, ", ")

			// Inform Libre
			for _, equip := range workCalendar.Equipment {
				s.LogDebugf("\tequipment: %s(%s): is currently %s with entries %s\n", equip.Name, equip.Id, calendarEntryType, calendarEntries)
				s.publishWorkCalendarType(equip, calendarEntryType)
				s.publishWorkCalendarEntryNames(equip, calendarEntries)
			}
		}
	}
}

func (s *calendarService) publishWorkCalendarType(equip domain.Equipment, calendarEntryType domain.WorkCalendarEntryType) {
	msg := domain.StdMessageStruct{
		OwningAsset:      equip.Name,
		OwningAssetId:    equip.Id,
		ItemName:         WorkCalendarCategory,
		ItemNameExt:      map[string]string{},
		ItemId:           "",
		ItemValue:        string(calendarEntryType),
		ItemDataType:     domain.DataTypeString,
		TagQuality:       1,
		Err:              nil,
		ChangedTimestamp: time.Now().UTC(),
		Category:         domain.SVCRQST_TAGDATA,
		Topic:            equip.Name + "/" + WorkCalendarCategory,
	}

	if lastState := s.cacheType[equip.Id]; lastState != calendarEntryType {
		msg.ItemOldValue = string(lastState)
		err := s.publish.SendStdMessage(msg)
		if err != nil {
			s.LogErrorf("failed to send message %v; got %s", msg, err)
		}
		s.cacheType[equip.Id] = calendarEntryType
	}
}

func (s *calendarService) publishWorkCalendarEntryNames(equip domain.Equipment, calendarEntry string) {
	msg := domain.StdMessageStruct{
		OwningAsset:      equip.Name,
		OwningAssetId:    equip.Id,
		ItemName:         WorkCalendarEntry,
		ItemNameExt:      map[string]string{},
		ItemId:           "",
		ItemValue:        calendarEntry,
		ItemDataType:     domain.DataTypeString,
		TagQuality:       1,
		Err:              nil,
		ChangedTimestamp: time.Now().UTC(),
		Category:         domain.SVCRQST_TAGDATA,
		Topic:            equip.Name + "/" + WorkCalendarEntry,
	}

	if lastState := s.cacheEntries[equip.Id]; lastState != calendarEntry {
		msg.ItemOldValue = string(lastState)
		err := s.publish.SendStdMessage(msg)
		if err != nil {
			s.LogErrorf("failed to send message %v; got %s", msg, err)
		}
		s.cacheEntries[equip.Id] = calendarEntry
	}
}
