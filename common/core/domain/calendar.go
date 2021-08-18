package domain

import (
	"errors"
	"fmt"
	"time"

	iso8601 "github.com/senseyeio/duration"
)

// WorkCalendarEntryType is the work type of a calendar entry. Planned Busy, Planned Shutdown or Planned Downtime.
type WorkCalendarEntryType string

const (
	// PlannedBusyTime represents that equipment should be executing/running
	PlannedBusyTime WorkCalendarEntryType = "PlannedBusyTime"

	// PlannedDowntime represents that equipment should NOT be executing/running due to planned activities
	PlannedDowntime WorkCalendarEntryType = "PlannedDowntime"

	// PlannedShutdown represents that equipment should NOT be executing/running due to being unscheduled
	PlannedShutdown WorkCalendarEntryType = "PlannedShutdown"
)

// CompareWorkCalendarEntryType takes two WorkCalendarEntryType's and applieds the business logic to determinute prescedant/dominant entry type.
// Used to determine the current WorkCalendarEntryType when multiple entries elapse a specified time and presecance is sought.
func CompareWorkCalendarEntryType(old WorkCalendarEntryType, new WorkCalendarEntryType) (changed bool, entryType WorkCalendarEntryType) {
	if old == new {
		return false, new
	}

	if old == PlannedBusyTime {
		return false, old
	}

	if new == PlannedBusyTime {
		return true, new
	}

	if new == PlannedDowntime {
		return true, new
	}

	return false, old
}

// Weekday is the day of the week represented by the first two characters of the day. E.g. 'MO', 'TU', etc.
type Weekday string

// AsTimeWeekday converts a local Weekday into a time library's Weekday
func (w Weekday) AsTimeWeekday() time.Weekday {
	switch w {
	case Monday:
		return time.Monday
	case Tuesday:
		return time.Tuesday
	case Wednesday:
		return time.Wednesday
	case Thursday:
		return time.Thursday
	case Friday:
		return time.Friday
	case Saturday:
		return time.Saturday
	}
	return time.Sunday // Default
}

// WeekdaySliceAsInts converts an array of Weekdays into an array of ints. Sunday = 0, Monday = 1...
func WeekdaySliceAsInts(arr []Weekday) (result []int) {
	for _, weekday := range arr {
		result = append(result, int(weekday.AsTimeWeekday()))
	}
	return result
}

const (
	// Monday GraphQL Enum Value
	Monday Weekday = "MO"

	// Tuesday GraphQL Enum Value
	Tuesday Weekday = "TU"

	// Wednesday GraphQL Enum Value
	Wednesday Weekday = "WE"

	// Thursday GraphQL Enum Value
	Thursday Weekday = "TH"

	// Friday GraphQL Enum Value
	Friday Weekday = "FR"

	// Saturday GraphQL Enum Value
	Saturday Weekday = "SA"

	// Sunday GraphQL Enum Value
	Sunday Weekday = "SU"
)

// Frequency of the WorkCalendarDefinition
type Frequency string

const (
	// Yearly Frequency of the WorkCalendarDefinition in GraphQL
	Yearly Frequency = "YEARLY"

	// Monthly Frequency of the WorkCalendarDefinition in GraphQL
	Monthly Frequency = "MONTHLY"

	// Weekly Frequency of the WorkCalendarDefinition in GraphQL
	Weekly Frequency = "WEEKLY"

	// Daily Frequency of the WorkCalendarDefinition in GraphQL
	Daily Frequency = "DAILY"

	// Hourly Frequency of the WorkCalendarDefinition in GraphQL
	Hourly Frequency = "HOURLY"

	// Minutely Frequency of the WorkCalendarDefinition in GraphQL
	Minutely Frequency = "MINUTELY"

	// Secondly Frequency of the WorkCalendarDefinition in GraphQL
	Secondly Frequency = "SECONDLY"
)

// WorkCalendar is a named collection of calendar entry definitions and entries that reflect the scheduled time blocks on equipment
type WorkCalendar struct {
	ID          string                        `json:"id,omitempty"`
	IsActive    bool                          `json:"isActive,omitempty"`
	Name        string                        `json:"name,omitempty"`
	Description string                        `json:"description,omitempty"`
	Definition  []WorkCalendarDefinitionEntry `json:"definition,omitempty"`
	Entries     []WorkCalendarEntry           `json:"entries,omitempty"`
	Equipment   []Equipment                   `json:"equipment,omitempty"`
}

// GetCurrentEntryType gets the current (now) WorkCalendarEntryType for a work calendar
func (workCalendar *WorkCalendar) GetCurrentEntryType() (entryType WorkCalendarEntryType, err error) {
	entryType = PlannedShutdown
	entries, err := workCalendar.GetCurrentEntries()
	if err != nil {
		msg := fmt.Sprintf("work calendar: %s(%s). failed to get calendar entries because %s", workCalendar.Name, workCalendar.ID, err)
		return entryType, errors.New(msg)
	}
	for _, entry := range entries {
		_, entryType = CompareWorkCalendarEntryType(entryType, entry.EntryType)
	}
	return entryType, nil
}

// GetCurrentEntries gets the all the WorkCalendarEntries that elapse over now for the WorkCalendar
func (workCalendar *WorkCalendar) GetCurrentEntries() (entries []WorkCalendarEntry, err error) {
	now := time.Now()
	return workCalendar.GetEntriesAtTime(now)
}

// GetEntriesAtTime gets the all the WorkCalendarEntries that elapse over a given time for the WorkCalendar
func (workCalendar *WorkCalendar) GetEntriesAtTime(atTime time.Time) (entries []WorkCalendarEntry, err error) {
	defEntries := []WorkCalendarEntry{}

	// Gather Entries
	for _, definition := range workCalendar.Definition {
		covers, err := definition.Covers(atTime)
		if err != nil {
			return entries, err
		}
		if covers {
			if defintionEntries, err := definition.GenerateEntries(); err == nil {
				defEntries = append(defEntries, defintionEntries...)
			} else {
				return entries, err
			}
		}
	}

	// Filter for any that cover time
	for _, defEntry := range defEntries {
		if atTime.Before(defEntry.EndDateTime) && (defEntry.StartDateTime.Before(atTime) || defEntry.StartDateTime.Equal(atTime)) {
			entries = append(entries, defEntry)
		}
	}
	return entries, nil
}

// WorkCalendarDefinitionEntry defintes a repeating pattern for workCalendarEntries
type WorkCalendarDefinitionEntry struct {
	ID          string `graphql:"id" json:"id,omitempty"`
	IsActive    bool   `graphql:"isActive" json:"isActive,omitempty"`
	Description string `graphql:"description" json:"description,omitempty"`
	// HierarchyScope Equipment `json:"hierarchyScope,omitempty"`
	Freq          Frequency             `json:"freq,omitempty"`
	StartDateTime time.Time             `json:"startDateTime,omitempty"`
	EndDateTime   time.Time             `json:"endDateTime,omitempty"`
	Count         int                   `json:"count,omitempty"`
	Interval      int                   `json:"interval,omitempty"`
	Weekday       Weekday               `graphql:"wkst" json:"wkst,omitempty"`
	ByWeekDay     []Weekday             `json:"Weekday,omitempty"`
	ByMonth       []int                 `json:"byMonth,omitempty"`
	BySetPos      []int                 `json:"bySetPos,omitempty"`
	ByMonthDay    []int                 `json:"byMonthDay,omitempty"`
	ByWeekNo      []int                 `json:"byWeekNo,omitempty"`
	ByHour        []int                 `json:"byHour,omitempty"`
	ByMinute      []int                 `json:"byMinute,omitempty"`
	BySecond      []int                 `json:"bySecond,omitempty"`
	ByYearDay     []int                 `json:"byYearDay,omitempty"`
	Duration      string                `json:"duration,omitempty"`
	EntryType     WorkCalendarEntryType `graphql:"entryType" json:"entryType,omitempty"`
}

// Covers checks if the given time is occurs within the scope of this WorkCalendarDefinitionEntry
func (def *WorkCalendarDefinitionEntry) Covers(time time.Time) (bool, error) {
	if !def.IsActive {
		return false, nil
	}

	if time.Before(def.StartDateTime) {
		return false, nil
	}

	if def.Count > 0 {
		return true, nil
	}

	definitionEnd, err := def.GetEndDateTime()
	return time.Before(definitionEnd), err
}

// GetEndDateTime returns/calulates the end date time of the WorkCalendarDefinitionEntry
func (def *WorkCalendarDefinitionEntry) GetEndDateTime() (time.Time, error) {
	if def.EndDateTime.IsZero() {
		return def.EndDateTime, errors.New("EndDateTime is zero")
	}
	return def.EndDateTime, nil
}

// GenerateEntries reads all the WorkCalendarDefinitions and creates the corresponding WorkCalendarEntries
func (def *WorkCalendarDefinitionEntry) GenerateEntries() ([]WorkCalendarEntry, error) {
	// Not Active No Entries
	if !def.IsActive {
		return []WorkCalendarEntry{}, nil
	}

	// Based on Frequeny
	switch def.Freq {
	case Daily:
		return def.getDailyEntries()
	case Weekly:
		return def.getWeeklyEntries()
	default:
		msg := fmt.Sprintf("WorkCalendarDefinitionEntry Frequency is undefined for %s", def.Description)
		return []WorkCalendarEntry{}, errors.New(msg)
	}
}

func intExists(arr []int, search int) bool {
	for _, i := range arr {
		if i == search {
			return true
		}
	}
	return false
}

type timeFilter struct {
	def          *WorkCalendarDefinitionEntry
	allSecond    bool
	allMinutes   bool
	allHours     bool
	allDays      bool
	allMonthDays bool
	allWeekNo    bool
	allMonths    bool
	allYearDays  bool
}

func (timeFilter *timeFilter) compare(time time.Time) bool {
	if !timeFilter.allSecond && !intExists(timeFilter.def.BySecond, time.Second()) {
		return false
	}

	if !timeFilter.allMinutes && !intExists(timeFilter.def.ByMinute, time.Minute()) {
		return false
	}

	if !timeFilter.allHours && !intExists(timeFilter.def.ByHour, time.Hour()) {
		return false
	}

	if !timeFilter.allMonthDays && !intExists(timeFilter.def.ByMonthDay, time.Day()) {
		return false
	}

	if !timeFilter.allMonths && !intExists(timeFilter.def.ByMonth, int(time.Month())) {
		return false
	}

	if !timeFilter.allYearDays && !intExists(timeFilter.def.ByYearDay, time.YearDay()) {
		return false
	}

	_, wk := time.ISOWeek()
	if !timeFilter.allWeekNo && !intExists(timeFilter.def.ByWeekNo, wk) {
		return false
	}

	weekDay := int(time.Weekday())
	byWeekDay := WeekdaySliceAsInts(timeFilter.def.ByWeekDay)

	return timeFilter.allDays || intExists(byWeekDay, weekDay)
}

func (def *WorkCalendarDefinitionEntry) getDailyEntries() (entries []WorkCalendarEntry, err error) {

	// Sanity Check on definition
	if def.Count == 0 && def.EndDateTime.IsZero() || def.StartDateTime.IsZero() {
		return entries, err
	}

	// Check Duration is valid
	dur, err := iso8601.ParseISO8601(def.Duration)
	if err != nil {
		return entries, err
	}

	tf := timeFilter{
		def:          def,
		allSecond:    len(def.BySecond) == 0,
		allMinutes:   len(def.ByMinute) == 0,
		allHours:     len(def.ByHour) == 0,
		allDays:      len(def.ByWeekDay) == 0,
		allMonthDays: len(def.ByMonthDay) == 0,
		allWeekNo:    len(def.ByWeekNo) == 0,
		allMonths:    len(def.ByMonth) == 0,
		allYearDays:  len(def.ByYearDay) == 0,
		// allSetPos: len(def.BySetPos) == 0 // No Idea what this is. T.H. 2021-08-13
	}

	start := def.StartDateTime

	sanityCheck := 10000 // Max out entries at this level

	if def.Count > 0 {
		counter := def.Count
		for counter > 0 && sanityCheck > 0 {
			sanityCheck--

			if !def.EndDateTime.IsZero() && def.EndDateTime.Before(start) {
				break
			}

			// Check Time Filter
			if !tf.compare(start) {
				start = start.AddDate(0, 0, 1)
				continue
			}

			entry := WorkCalendarEntry{
				ID:            "",
				IsActive:      true,
				Description:   def.Description,
				StartDateTime: start,
				EndDateTime:   dur.Shift(start),
				EntryType:     def.EntryType,
			}
			entries = append(entries, entry)
			counter--
			start = start.AddDate(0, 0, 1)
		}
	}

	if sanityCheck < 0 {
		err = errors.New("suspect WorkCalendarDefintion configuration error, walked over 10000 times trying to find next WorkCalendarEntry")
	}

	return entries, err
}

func (def *WorkCalendarDefinitionEntry) getWeeklyEntries() (entries []WorkCalendarEntry, err error) {
	outerTime := def.StartDateTime
	count := 0

	dur, err := iso8601.ParseISO8601(def.Duration)
	if err != nil {
		return entries, err
	}

	hour := 0
	minute := 0
	second := 0
	if len(def.ByHour) > 0 {
		hour = def.ByHour[0] // Just use first for now
	}

	if len(def.ByMinute) > 0 {
		minute = def.ByMinute[0] // Just use first for now
	}

	if len(def.BySecond) > 0 {
		second = def.BySecond[0] // Just use first for now
	}

	// Go ahead until weekly start date
	for outerTime.Weekday() != def.Weekday.AsTimeWeekday() {
		outerTime = outerTime.AddDate(0, 0, 1)
	}

	weeklyEndTime := outerTime.AddDate(0, 0, 7)

	for weeklyEndTime.Before(def.EndDateTime) && (def.Count == 0 || count < def.Count) {
		for _, day := range def.ByWeekDay {

			// This day is technically in the following week
			days := int(day.AsTimeWeekday()) - int(def.Weekday.AsTimeWeekday())
			if days < 0 {
				days = days + 7
			}
			startDateTime := outerTime.AddDate(0, 0, days)
			startDateTime = time.Date(startDateTime.Year(), startDateTime.Month(), startDateTime.Day(), hour, minute, second, 0, time.UTC)
			entry := WorkCalendarEntry{
				ID:            "",
				IsActive:      true,
				Description:   def.Description,
				StartDateTime: startDateTime,
				EndDateTime:   dur.Shift(startDateTime),
				EntryType:     def.EntryType,
			}
			entries = append(entries, entry)
			count++
		}
		outerTime = outerTime.AddDate(0, 0, 7)
		weeklyEndTime = outerTime.AddDate(0, 0, 7)
	}

	return entries, nil
}

// WorkCalendarEntry is a block of time with start/end, label and WorkCalendarEntryType
type WorkCalendarEntry struct {
	ID            string    `json:"id,omitempty"`
	IsActive      bool      `json:"isActive,omitempty"`
	Description   string    `json:"description,omitempty"`
	StartDateTime time.Time `json:"startDateTime,omitempty"`
	EndDateTime   time.Time `json:"endDateTime,omitempty"`
	EntryType     WorkCalendarEntryType
}
