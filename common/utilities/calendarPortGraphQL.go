package utilities

import (
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-common/common/core/queries"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
)

type calendarPortGraphQL struct {
	//inherit config functions
	libreConfig.ConfigurationEnabler

	//inherit logging functions
	libreLogger.LoggingEnabler

	dataStore ports.LibreDataStorePort
}

// NewCalendarPortGraphQL creats and bootstraps a new Calendar Port for GraphQL
func NewCalendarPortGraphQL(configHook string, storeIF ports.LibreDataStorePort) *calendarPortGraphQL {
	s := calendarPortGraphQL{
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

func (s *calendarPortGraphQL) GetAllActiveWorkCalendar() ([]domain.WorkCalendar, error) {
	txn := s.dataStore.BeginTransaction(false, "findActiveWorkCalendar")
	defer txn.Dispose()

	workCalendars, err := queries.GetAllActiveWorkCalendar(txn)
	return workCalendars, err
}
