package ports

import "github.com/Spruik/libre-common/common/core/domain"

//The CalendarPort interface defines the functions to support the getting calendars
type CalendarPort interface {

	//GetAllWorkCalendar gets all the work calendars
	GetAllActiveWorkCalendar() ([]domain.WorkCalendar, error)
}
