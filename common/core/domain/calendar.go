package domain

import (
	"errors"
	"fmt"
	"time"

	iso8601 "github.com/senseyeio/duration"
)

type WorkCalendarEntryType string

const (
	PlannedBusyTime WorkCalendarEntryType = "PlannedBusyTime"
	PlannedDowntime WorkCalendarEntryType = "PlannedDowntime"
	PlannedShutdown WorkCalendarEntryType = "PlannedShutdown"
)

func DominantWorkCalendarEntryType(old WorkCalendarEntryType, new WorkCalendarEntryType) (changed bool, entryType WorkCalendarEntryType) {
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

type Weekday string

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

func WeekdaySliceAsInts(arr []Weekday) (result []int) {
	for _, weekday := range arr {
		result = append(result, int(weekday.AsTimeWeekday()))
	}
	return result
}

const (
	Monday    Weekday = "MO"
	Tuesday   Weekday = "TU"
	Wednesday Weekday = "WE"
	Thursday  Weekday = "TH"
	Friday    Weekday = "FR"
	Saturday  Weekday = "SA"
	Sunday    Weekday = "SU"
)

type Frequency string

const (
	Yearly   Frequency = "YEARLY"
	Monthly  Frequency = "MONTHLY"
	Weekly   Frequency = "WEEKLY"
	Daily    Frequency = "DAILY"
	Hourly   Frequency = "HOURLY"
	Minutely Frequency = "MINUTELY"
	Secondly Frequency = "SECONDLY"
)

type WorkCalendar struct {
	Id          string                        `json:"id,omitempty"`
	IsActive    bool                          `json:"isActive,omitempty"`
	Name        string                        `json:"name,omitempty"`
	Description string                        `json:"description,omitempty"`
	Definition  []WorkCalendarDefinitionEntry `json:"definition,omitempty"`
	Entries     []WorkCalendarEntry           `json:"entries,omitempty"`
	Equipment   []Equipment                   `json:"equipment,omitempty"`
}

func (calendar *WorkCalendar) GetCurrentEntryType() (entryType WorkCalendarEntryType, err error) {
	entryType = PlannedShutdown
	entries, err := calendar.GetCurrentEntries()
	if err != nil {
		msg := fmt.Sprintf("work calendar: %s(%s). failed to get calendar entries due to %s", calendar.Name, calendar.Id, err)
		return entryType, errors.New(msg)
	}
	for _, entry := range entries {
		_, entryType = DominantWorkCalendarEntryType(entryType, entry.EntryType)
	}
	return entryType, nil
}

func (calendar *WorkCalendar) GetCurrentEntries() (entries []WorkCalendarEntry, err error) {
	now := time.Now()
	return calendar.GetEntriesAtTime(now)
}

func (workCalendar *WorkCalendar) GetEntriesAtTime(time time.Time) (entries []WorkCalendarEntry, err error) {
	defEntries := []WorkCalendarEntry{}

	// Gather Entries
	for _, definition := range workCalendar.Definition {
		covers, err := definition.Covers(time)
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
		if defEntry.StartDateTime.Before(time) && time.Before(defEntry.EndDateTime) {
			entries = append(entries, defEntry)
		}
	}
	return entries, nil
}

type WorkCalendarDefinitionEntry struct {
	Id          string `graphql:"id" json:"id,omitempty"`
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
	// CalendarEntries []WorkCalendarEntry `json:"calendarEntries,omitempty"`	// Ignoring for now T.H. 2021-08-12
	// WorkCalendar    *WorkCalendar       `json:"workCalendar,omitempty"`		// Ignoring for now T.H. 2021-08-12
	//properties:[Property] 													// Ignoring for now T.H. 2021-08-12
}

func (definition *WorkCalendarDefinitionEntry) Covers(time time.Time) (bool, error) {
	if !definition.IsActive {
		return false, nil
	}

	if time.Before(definition.StartDateTime) {
		return false, nil
	}

	if definition.Count > 0 {
		return true, nil
	}

	definitionEnd, err := definition.GetEndDateTime()
	return time.Before(definitionEnd), err
}

func (def *WorkCalendarDefinitionEntry) GetEndDateTime() (time.Time, error) {
	if def.EndDateTime.IsZero() {
		return def.EndDateTime, errors.New("EndDateTime is zero")
	}
	return def.EndDateTime, nil
}

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

func (timeFilter *timeFilter) Compare(time time.Time) bool {
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

	if def.Count > 0 {
		counter := def.Count
		for counter > 0 {
			// Check Time Filter
			if !tf.Compare(start) {
				start = start.AddDate(0, 0, 1)
				continue
			}

			entry := WorkCalendarEntry{
				Id:            "",
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
		fmt.Printf("\t\tEvaluating Week %s to %s\n", outerTime.Format(time.RFC3339), weeklyEndTime.Format(time.RFC3339))
		for _, day := range def.ByWeekDay {

			// This day is technically in the following week
			days := int(day.AsTimeWeekday()) - int(def.Weekday.AsTimeWeekday())
			if days < 0 {
				days = days + 7
			}
			startDateTime := outerTime.AddDate(0, 0, days)
			startDateTime = time.Date(startDateTime.Year(), startDateTime.Month(), startDateTime.Day(), hour, minute, second, 0, time.UTC)
			entry := WorkCalendarEntry{
				Id:            "",
				IsActive:      true,
				Description:   def.Description,
				StartDateTime: startDateTime,
				EndDateTime:   dur.Shift(startDateTime),
				EntryType:     def.EntryType,
			}
			fmt.Printf("Adding Entry: %v\n", entry)
			entries = append(entries, entry)
			count++
		}
		outerTime = outerTime.AddDate(0, 0, 7)
		weeklyEndTime = outerTime.AddDate(0, 0, 7)
	}

	return entries, nil
}

type WorkCalendarEntry struct {
	Id          string `json:"id,omitempty"`
	IsActive    bool   `json:"isActive,omitempty"`
	Description string `json:"description,omitempty"`
	// HierarchyScope Equipment                    `json:"hierarchyScope,omitempty"`
	// Definition    *WorkCalendarDefinitionEntry `json:"definition,omitempty"`
	StartDateTime time.Time `json:"startDateTime,omitempty"`
	EndDateTime   time.Time `json:"endDateTime,omitempty"`
	EntryType     WorkCalendarEntryType
	//properties:[Property] // Ignoring for now T.H. 2021-08-12
	// WorkCalendar *WorkCalendar `json:"workCalendar,omitempty"`
}
