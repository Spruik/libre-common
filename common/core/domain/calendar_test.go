package domain

import (
	"fmt"
	"testing"
	"time"
)

func mustMakeTime(str string) (t time.Time) {
	t, _ = time.Parse(time.RFC3339, str)
	return t
}

type weekdaySliceAsIntsTestCase struct {
	Name     string
	Weekdays []Weekday
	Expected []int
}

var weekdaySliceAsIntsTestCases = []weekdaySliceAsIntsTestCase{
	{
		Name:     "Simple WeekdaySliceAsInts Test",
		Weekdays: []Weekday{Monday, Tuesday, Thursday, Sunday},
		Expected: []int{1, 2, 4, 0},
	},
}

func TestWeekdaySliceAsIntsTestCases(t *testing.T) {
	for _, tc := range weekdaySliceAsIntsTestCases {
		result := WeekdaySliceAsInts(tc.Weekdays)

		for i := range result {
			if result[i] != tc.Expected[i] {
				t.Errorf("Test Case '%s': at index %d got %d; want %d", tc.Name, i, result[i], tc.Expected[i])
			}
		}
	}

	t.Log("Complete TestWeekdaySliceAsIntsTestCases")
}

type compareWorkCalendarEntryTypeTestCase struct {
	Name     string
	Old      WorkCalendarEntryType
	New      WorkCalendarEntryType
	Expected WorkCalendarEntryType
	Changed  bool
}

// Test cases to compare two WorkCalendarEntryTypes to determine presedence
var compareWorkCalendarEntryTypeTestCases = []compareWorkCalendarEntryTypeTestCase{
	{
		Name:     "Same",
		Old:      PlannedShutdown,
		New:      PlannedShutdown,
		Expected: PlannedShutdown,
		Changed:  false,
	},
	{
		Name:     "Transition PlannedShutdown to PlannedBusytime",
		Old:      PlannedShutdown,
		New:      PlannedBusyTime,
		Expected: PlannedBusyTime,
		Changed:  true,
	},
	{
		Name:     "Transition PlannedBusytime to PlannedShutdown",
		Old:      PlannedBusyTime,
		New:      PlannedShutdown,
		Expected: PlannedBusyTime,
		Changed:  false,
	},
	{
		Name:     "Transition PlannedShutdown to PlannedDowntime",
		Old:      PlannedShutdown,
		New:      PlannedDowntime,
		Expected: PlannedDowntime,
		Changed:  true,
	},
	{
		Name:     "Transition PlannedDowntime to PlannedBusyTime",
		Old:      PlannedDowntime,
		New:      PlannedBusyTime,
		Expected: PlannedBusyTime,
		Changed:  true,
	},
	{
		Name:     "Transition PlannedDowntime to PlannedBusyTime",
		Old:      PlannedBusyTime,
		New:      PlannedDowntime,
		Expected: PlannedBusyTime,
		Changed:  false,
	},
}

func TestCompareWorkCalendarEntryType(t *testing.T) {
	for _, tc := range compareWorkCalendarEntryTypeTestCases {
		changed, result := CompareWorkCalendarEntryType(tc.Old, tc.New)

		if result != tc.Expected {
			t.Errorf("Test Case '%s': got %s; want %s", tc.Name, result, tc.Expected)
		}

		if changed != tc.Changed {
			t.Errorf("Test Case '%s': got %t; want %t", tc.Name, changed, tc.Changed)
		}
	}

	t.Log("Complete TestCompareWorkCalendarEntryType")
}

type entryTestCase struct {
	Name         string
	WorkCalendar WorkCalendar
	Now          time.Time
	Entries      []WorkCalendarEntry
	Error        bool
}

var entryTestCases = []entryTestCase{
	{
		Name: "WorkCalendar No Definitions",
		WorkCalendar: WorkCalendar{
			ID:          "abc123",
			IsActive:    true,
			Name:        "Work Calendar - No Definitions",
			Description: "Description",
			Entries:     []WorkCalendarEntry{},
			Equipment:   []Equipment{},
			Definition:  []WorkCalendarDefinitionEntry{},
		},
		Now:     mustMakeTime("2021-08-13T00:00:00Z"),
		Entries: []WorkCalendarEntry{},
		Error:   false,
	},
	{
		Name: "WorkCalendar Single Definition",
		WorkCalendar: WorkCalendar{
			ID:          "abc123",
			IsActive:    true,
			Name:        "Work Calendar - Single Definition",
			Description: "Description",
			Entries:     []WorkCalendarEntry{},
			Equipment:   []Equipment{},
			Definition: []WorkCalendarDefinitionEntry{
				{
					ID:            "abc123",
					IsActive:      true,
					Description:   "Shift A",
					Freq:          Weekly,
					StartDateTime: mustMakeTime("2021-01-01T00:00:00Z"),
					EndDateTime:   mustMakeTime("2022-01-01T00:00:00Z"),
					Count:         365,
					Interval:      0,
					Weekday:       Monday,
					ByWeekDay:     []Weekday{Monday, Tuesday, Wednesday, Thursday, Friday},
					ByMonth:       []int{},
					BySetPos:      []int{},
					ByMonthDay:    []int{},
					ByWeekNo:      []int{},
					ByHour:        []int{8},
					ByMinute:      []int{0},
					BySecond:      []int{0},
					ByYearDay:     []int{},
					Duration:      "PT8H",
					EntryType:     PlannedBusyTime,
				},
			},
		},
		Now: mustMakeTime("2021-08-13T12:00:00Z"),
		Entries: []WorkCalendarEntry{
			{
				ID:            "abc123",
				IsActive:      true,
				Description:   "Shift A",
				StartDateTime: mustMakeTime("2021-08-13T08:00:00Z"),
				EndDateTime:   mustMakeTime("2021-08-13T16:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
		},
		Error: false,
	},
	{
		Name: "WorkCalendar Single Definition query before",
		WorkCalendar: WorkCalendar{
			ID:          "abc123",
			IsActive:    true,
			Name:        "Work Calendar - Single Definition",
			Description: "Description",
			Entries:     []WorkCalendarEntry{},
			Equipment:   []Equipment{},
			Definition: []WorkCalendarDefinitionEntry{
				{
					ID:            "abc123",
					IsActive:      true,
					Description:   "Shift A",
					Freq:          Weekly,
					StartDateTime: mustMakeTime("2021-01-01T00:00:00Z"),
					EndDateTime:   mustMakeTime("2022-01-01T00:00:00Z"),
					Count:         365,
					Interval:      0,
					Weekday:       Monday,
					ByWeekDay:     []Weekday{Monday, Tuesday, Wednesday, Thursday, Friday},
					ByMonth:       []int{},
					BySetPos:      []int{},
					ByMonthDay:    []int{},
					ByWeekNo:      []int{},
					ByHour:        []int{8},
					ByMinute:      []int{0},
					BySecond:      []int{0},
					ByYearDay:     []int{},
					Duration:      "PT8H",
					EntryType:     PlannedBusyTime,
				},
			},
		},
		Now:     mustMakeTime("2020-01-01T01:00:00Z"),
		Entries: []WorkCalendarEntry{},
		Error:   false,
	},
	{
		Name: "WorkCalendar Single Definition query after",
		WorkCalendar: WorkCalendar{
			ID:          "abc123",
			IsActive:    true,
			Name:        "Work Calendar - Single Definition",
			Description: "Description",
			Entries:     []WorkCalendarEntry{},
			Equipment:   []Equipment{},
			Definition: []WorkCalendarDefinitionEntry{
				{
					ID:            "abc123",
					IsActive:      true,
					Description:   "Shift A",
					Freq:          Daily,
					StartDateTime: mustMakeTime("2021-01-01T00:00:00Z"),
					EndDateTime:   mustMakeTime("2022-01-01T00:00:00Z"),
					Count:         365,
					Interval:      0,
					Weekday:       Monday,
					ByWeekDay:     []Weekday{Monday, Tuesday, Wednesday, Thursday, Friday},
					ByMonth:       []int{},
					BySetPos:      []int{},
					ByMonthDay:    []int{},
					ByWeekNo:      []int{},
					ByHour:        []int{8},
					ByMinute:      []int{0},
					BySecond:      []int{0},
					ByYearDay:     []int{},
					Duration:      "PT8H",
					EntryType:     PlannedBusyTime,
				},
			},
		},
		Now:     mustMakeTime("2023-01-01T01:00:00Z"),
		Entries: []WorkCalendarEntry{},
		Error:   false,
	},
	{
		Name: "WorkCalendar Single Definition - Time aligns at start of first calendar entry",
		WorkCalendar: WorkCalendar{
			ID:          "abc123",
			IsActive:    true,
			Name:        "Work Calendar - Single Definition",
			Description: "Description",
			Entries:     []WorkCalendarEntry{},
			Equipment:   []Equipment{},
			Definition: []WorkCalendarDefinitionEntry{
				{
					ID:            "abc123",
					IsActive:      true,
					Description:   "Shift A",
					Freq:          Weekly,
					StartDateTime: mustMakeTime("2021-01-04T08:00:00Z"),
					EndDateTime:   mustMakeTime("2022-01-01T00:00:00Z"),
					Count:         365,
					Interval:      0,
					Weekday:       Monday,
					ByWeekDay:     []Weekday{Monday, Tuesday, Wednesday, Thursday, Friday},
					ByMonth:       []int{},
					BySetPos:      []int{},
					ByMonthDay:    []int{},
					ByWeekNo:      []int{},
					ByHour:        []int{8},
					ByMinute:      []int{0},
					BySecond:      []int{0},
					ByYearDay:     []int{},
					Duration:      "PT8H",
					EntryType:     PlannedBusyTime,
				},
			},
		},
		Now: mustMakeTime("2021-01-04T08:00:00Z"),
		Entries: []WorkCalendarEntry{
			{
				ID:            "abc123",
				IsActive:      true,
				Description:   "Shift A",
				StartDateTime: mustMakeTime("2021-01-04T08:00:00Z"),
				EndDateTime:   mustMakeTime("2021-01-04T16:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
		},
		Error: false,
	},
	{
		Name: "WorkCalendar Single Definition - Time aligns at end of first calendar entry",
		WorkCalendar: WorkCalendar{
			ID:          "abc123",
			IsActive:    true,
			Name:        "Work Calendar - Single Definition",
			Description: "Description",
			Entries:     []WorkCalendarEntry{},
			Equipment:   []Equipment{},
			Definition: []WorkCalendarDefinitionEntry{
				{
					ID:            "abc123",
					IsActive:      true,
					Description:   "Shift A",
					Freq:          Weekly,
					StartDateTime: mustMakeTime("2021-01-04T08:00:00Z"),
					EndDateTime:   mustMakeTime("2022-01-01T00:00:00Z"),
					Count:         365,
					Interval:      0,
					Weekday:       Monday,
					ByWeekDay:     []Weekday{Monday, Tuesday, Wednesday, Thursday, Friday},
					ByMonth:       []int{},
					BySetPos:      []int{},
					ByMonthDay:    []int{},
					ByWeekNo:      []int{},
					ByHour:        []int{8},
					ByMinute:      []int{0},
					BySecond:      []int{0},
					ByYearDay:     []int{},
					Duration:      "PT8H",
					EntryType:     PlannedBusyTime,
				},
			},
		},
		Now:     mustMakeTime("2021-01-04T16:00:00Z"),
		Entries: []WorkCalendarEntry{},
		Error:   false,
	},
	{
		Name: "WorkCalendar Daily with EndDateTIme",
		WorkCalendar: WorkCalendar{
			ID:          "abc123",
			IsActive:    true,
			Name:        "Work Calendar - Daily with EndDateTIme",
			Description: "Description",
			Entries:     []WorkCalendarEntry{},
			Equipment:   []Equipment{},
			Definition: []WorkCalendarDefinitionEntry{
				{
					ID:            "abc123",
					IsActive:      true,
					Description:   "Daily Shift",
					Freq:          Daily,
					StartDateTime: mustMakeTime("2021-01-01T08:00:00Z"),
					EndDateTime:   mustMakeTime("2021-01-07T08:00:00Z"),
					Count:         0,
					Interval:      0,
					Weekday:       Monday,
					ByWeekDay:     []Weekday{Monday, Tuesday, Wednesday, Thursday, Friday},
					ByMonth:       []int{},
					BySetPos:      []int{},
					ByMonthDay:    []int{},
					ByWeekNo:      []int{},
					ByHour:        []int{},
					ByMinute:      []int{},
					BySecond:      []int{},
					ByYearDay:     []int{},
					Duration:      "PT8H",
					EntryType:     PlannedBusyTime,
				},
			},
		},
		Now: mustMakeTime("2021-01-04T12:00:00Z"),
		Entries: []WorkCalendarEntry{
			{
				ID:            "abc123",
				IsActive:      true,
				Description:   "Daily Shift",
				StartDateTime: mustMakeTime("2021-01-04T08:00:00Z"),
				EndDateTime:   mustMakeTime("2021-01-04T16:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
		},
		Error: false,
	},
	{
		Name: "WorkCalendar Daily with Count",
		WorkCalendar: WorkCalendar{
			ID:          "abc123",
			IsActive:    true,
			Name:        "Work Calendar - Daily with EndDateTIme",
			Description: "Description",
			Entries:     []WorkCalendarEntry{},
			Equipment:   []Equipment{},
			Definition: []WorkCalendarDefinitionEntry{
				{
					ID:            "abc123",
					IsActive:      true,
					Description:   "Daily Shift",
					Freq:          Daily,
					StartDateTime: mustMakeTime("2021-01-01T08:00:00Z"),
					EndDateTime:   mustMakeTime("0001-01-01T00:00:00Z"),
					Count:         20,
					Interval:      0,
					Weekday:       Monday,
					ByWeekDay:     []Weekday{Monday, Tuesday, Wednesday, Thursday, Friday},
					ByMonth:       []int{},
					BySetPos:      []int{},
					ByMonthDay:    []int{},
					ByWeekNo:      []int{},
					ByHour:        []int{},
					ByMinute:      []int{},
					BySecond:      []int{},
					ByYearDay:     []int{},
					Duration:      "PT8H",
					EntryType:     PlannedBusyTime,
				},
			},
		},
		Now: mustMakeTime("2021-01-06T12:12:12Z"),
		Entries: []WorkCalendarEntry{
			{
				ID:            "abc123",
				IsActive:      true,
				Description:   "Daily Shift",
				StartDateTime: mustMakeTime("2021-01-06T08:00:00Z"),
				EndDateTime:   mustMakeTime("2021-01-06T16:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
		},
		Error: false,
	},
}

func TestWorkCalendarGetEntriesAtTime(t *testing.T) {
	for _, tc := range entryTestCases {
		entries, err := tc.WorkCalendar.GetEntriesAtTime(tc.Now)

		// Expect Error
		if tc.Error && err == nil {
			t.Errorf("Test Case '%s': got no error; want error", tc.Name)
		}

		// Don't Expect Error
		if !tc.Error && err != nil {
			t.Errorf("Test Case '%s': got error: %s; want no error", tc.Name, err)
		}

		// Expect Same Entry Count
		if len(entries) != len(tc.Entries) {
			t.Errorf("Test Case '%s': got %d entries; want %d", tc.Name, len(entries), len(tc.Entries))
		}

		// Expect Same Entries
		for _, expectedEntry := range tc.Entries {
			found := false
			for _, actualEntry := range entries {
				if expectedEntry.Description == actualEntry.Description && expectedEntry.StartDateTime.Equal(actualEntry.StartDateTime) && expectedEntry.EndDateTime.Equal(actualEntry.EndDateTime) && actualEntry.EntryType == expectedEntry.EntryType {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Test Case '%s': expected to find entry %v; but didn't", tc.Name, expectedEntry)
			}
		}
	}

	t.Log("Complete TestWorkCalendarGetEntries")
}

type getEntriesAtTimeTestCase struct {
	Name         string
	WorkCalendar WorkCalendar
	Entries      []WorkCalendarEntry
	Error        bool
}

var getEntriesAtTimeTestCases = []getEntriesAtTimeTestCase{
	{
		Name: "No Definitions - Empty Entries",
		WorkCalendar: WorkCalendar{
			ID:          "test",
			IsActive:    true,
			Name:        "Test",
			Description: "Test GetCurrentEntryType",
			Definition:  []WorkCalendarDefinitionEntry{},
			Entries:     []WorkCalendarEntry{},
			Equipment:   []Equipment{},
		},
		Entries: []WorkCalendarEntry{},
		Error:   false,
	},
	{
		Name: "One Definition, One Entry",
		WorkCalendar: WorkCalendar{
			ID:          "0x123",
			IsActive:    true,
			Name:        "Test",
			Description: "Test GetCurrentEntryType",
			Definition: []WorkCalendarDefinitionEntry{
				{
					ID:            "0x124",
					IsActive:      true,
					Description:   "Shift A",
					Freq:          Daily,
					StartDateTime: mustMakeTime("2021-08-16T08:00:00Z"),
					EndDateTime:   mustMakeTime("2021-08-17T08:00:00Z"),
					Count:         1,
					Interval:      1,
					Weekday:       Monday,
					ByWeekDay:     []Weekday{Monday},
					ByMonth:       []int{},
					BySetPos:      []int{},
					ByMonthDay:    []int{},
					ByWeekNo:      []int{},
					ByHour:        []int{},
					ByMinute:      []int{},
					BySecond:      []int{},
					ByYearDay:     []int{},
					Duration:      "PT8H",
					EntryType:     PlannedBusyTime,
				},
			},
			Entries:   []WorkCalendarEntry{},
			Equipment: []Equipment{},
		},
		Entries: []WorkCalendarEntry{
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A",
				StartDateTime: mustMakeTime("2021-08-16T08:00:00Z"),
				EndDateTime:   mustMakeTime("2021-08-16T16:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
		},
		Error: false,
	},
}

func TestGetEntriesAtTime(t *testing.T) {
	for _, tc := range getEntriesAtTimeTestCases {
		entries, err := tc.WorkCalendar.GetEntriesAtTime(mustMakeTime("2021-08-16T11:00:00Z"))

		if tc.Error && err == nil || !tc.Error && err != nil {
			var expected string
			if tc.Error {
				expected = " "
			} else {
				expected = "no "
			}
			var got string
			if err == nil {
				got = "no error"
			} else {
				got = fmt.Sprintf("%s", err)
			}
			t.Errorf("Test Case '%s': expected %serror, got %s", tc.Name, expected, got)
		}

		if len(entries) != len(tc.Entries) {
			t.Errorf("Test Case '%s': expected %d entries, got %d", tc.Name, len(tc.Entries), len(entries))
		} else {
			for i := range entries {
				if entries[i] != tc.Entries[i] {
					t.Errorf("Test Case '%s': index %d want %v; got %v", tc.Name, i, tc.Entries[i], entries[i])
				}
			}
		}
	}

	t.Log("Complete TestGetCurrentEntryType")
}

func TestGetCurrentEntryType(t *testing.T) {
	now := time.Now().UTC()
	var wkd Weekday
	switch now.Weekday() {
	case time.Sunday:
		wkd = Sunday
	case time.Monday:
		wkd = Monday
	case time.Tuesday:
		wkd = Tuesday
	case time.Wednesday:
		wkd = Wednesday
	case time.Thursday:
		wkd = Thursday
	case time.Friday:
		wkd = Friday
	case time.Saturday:
		wkd = Saturday
	}
	workCalendar := WorkCalendar{
		ID:          "test",
		Name:        "Test",
		IsActive:    true,
		Description: "Test Work Calendar",
		Definition: []WorkCalendarDefinitionEntry{
			{
				ID:            "Test",
				IsActive:      true,
				Description:   "test",
				Freq:          Daily,
				StartDateTime: now.Add(-1 * 24 * time.Hour),
				EndDateTime:   now.Add(30 * 24 * time.Hour),
				Count:         35,
				Interval:      1,
				Weekday:       wkd,
				ByWeekDay:     []Weekday{},
				ByMonth:       []int{},
				BySetPos:      []int{},
				ByMonthDay:    []int{},
				ByWeekNo:      []int{},
				ByHour:        []int{},
				ByMinute:      []int{},
				BySecond:      []int{},
				ByYearDay:     []int{},
				Duration:      "PT8H",
				EntryType:     PlannedBusyTime,
			},
		},
		Entries:   []WorkCalendarEntry{},
		Equipment: []Equipment{},
	}

	entryType, names, err := workCalendar.GetCurrentEntryTypeAndNames()

	if err != nil {
		t.Errorf("Failed to GetCurrentEntryType. Expected no error got %s", err)
	}

	if entryType != PlannedBusyTime {
		t.Errorf("Failed to GetCurrentEntryType")
	}

	for _, name := range names {
		if name != "test" {
			t.Errorf("Failed to get names of GetCurrentEntryType")
		}
	}
}

var testCasesGetEntries = []entryTestCase{
	{
		Name: "Fortnightly Shift Pattern 12H Shifts, 3-2-2-3-2-2",
		WorkCalendar: WorkCalendar{
			ID:          "0x456",
			IsActive:    true,
			Name:        "Real Example",
			Description: "Fortnightly Shift Pattern 12H Shifts, 3-2-2-3-2-2",
			Definition: []WorkCalendarDefinitionEntry{
				{
					ID:            "0x124",
					IsActive:      true,
					Description:   "Shift A - Odd",
					Freq:          Daily,
					StartDateTime: mustMakeTime("2021-08-30T11:00:00Z"),
					EndDateTime:   mustMakeTime("2021-10-04T11:00:00Z"),
					Count:         0,
					Interval:      0,
					Weekday:       Sunday,
					ByWeekDay:     []Weekday{Monday, Saturday, Wednesday, Tuesday},
					ByMonth:       []int{},
					BySetPos:      []int{},
					ByMonthDay:    []int{},
					ByWeekNo:      []int{43, 51, 49, 47, 41, 39, 35, 37, 45},
					ByHour:        []int{},
					ByMinute:      []int{},
					BySecond:      []int{},
					ByYearDay:     []int{},
					Duration:      "PT12H",
					EntryType:     PlannedBusyTime,
				},
				{
					ID:            "0x125",
					IsActive:      true,
					Description:   "Shift A - Even",
					Freq:          Daily,
					StartDateTime: mustMakeTime("2021-08-30T11:00:00Z"),
					EndDateTime:   mustMakeTime("2021-10-04T11:00:00Z"),
					Count:         0,
					Interval:      0,
					Weekday:       Sunday,
					ByWeekDay:     []Weekday{Thursday, Sunday, Friday},
					ByMonth:       []int{},
					BySetPos:      []int{},
					ByMonthDay:    []int{},
					ByWeekNo:      []int{38, 40, 50, 48, 46, 44, 52, 42, 36},
					ByHour:        []int{},
					ByMinute:      []int{},
					BySecond:      []int{},
					ByYearDay:     []int{},
					Duration:      "PT12H",
					EntryType:     PlannedBusyTime,
				},
			},
			Entries:   []WorkCalendarEntry{},
			Equipment: []Equipment{},
		},
		Now: mustMakeTime("2021-09-15T12:12:12Z"),
		Entries: []WorkCalendarEntry{
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A - Odd",
				StartDateTime: mustMakeTime("2021-08-30T11:00:00Z"),
				EndDateTime:   mustMakeTime("2021-08-30T23:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A - Odd",
				StartDateTime: mustMakeTime("2021-08-31T11:00:00Z"),
				EndDateTime:   mustMakeTime("2021-08-31T23:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A - Odd",
				StartDateTime: mustMakeTime("2021-09-01T11:00:00Z"),
				EndDateTime:   mustMakeTime("2021-09-01T23:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
			// Off for two
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A - Odd",
				StartDateTime: mustMakeTime("2021-09-04T11:00:00Z"),
				EndDateTime:   mustMakeTime("2021-09-04T23:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A - Even",
				StartDateTime: mustMakeTime("2021-09-05T11:00:00Z"),
				EndDateTime:   mustMakeTime("2021-09-05T23:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
			// Off for three
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A - Even",
				StartDateTime: mustMakeTime("2021-09-09T11:00:00Z"),
				EndDateTime:   mustMakeTime("2021-09-09T23:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A - Even",
				StartDateTime: mustMakeTime("2021-09-10T11:00:00Z"),
				EndDateTime:   mustMakeTime("2021-09-10T23:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
			// Off for two
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A - Odd",
				StartDateTime: mustMakeTime("2021-09-13T11:00:00Z"),
				EndDateTime:   mustMakeTime("2021-09-13T23:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A - Odd",
				StartDateTime: mustMakeTime("2021-09-14T11:00:00Z"),
				EndDateTime:   mustMakeTime("2021-09-14T23:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A - Odd",
				StartDateTime: mustMakeTime("2021-09-15T11:00:00Z"),
				EndDateTime:   mustMakeTime("2021-09-15T23:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
			// off for two
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A - Odd",
				StartDateTime: mustMakeTime("2021-09-18T11:00:00Z"),
				EndDateTime:   mustMakeTime("2021-09-18T23:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A - Even",
				StartDateTime: mustMakeTime("2021-09-19T11:00:00Z"),
				EndDateTime:   mustMakeTime("2021-09-19T23:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
			// off for three
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A - Even",
				StartDateTime: mustMakeTime("2021-09-23T11:00:00Z"),
				EndDateTime:   mustMakeTime("2021-09-23T23:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A - Even",
				StartDateTime: mustMakeTime("2021-09-24T11:00:00Z"),
				EndDateTime:   mustMakeTime("2021-09-24T23:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
			// off for two
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A - Odd",
				StartDateTime: mustMakeTime("2021-09-27T11:00:00Z"),
				EndDateTime:   mustMakeTime("2021-09-27T23:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A - Odd",
				StartDateTime: mustMakeTime("2021-09-28T11:00:00Z"),
				EndDateTime:   mustMakeTime("2021-09-28T23:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A - Odd",
				StartDateTime: mustMakeTime("2021-09-29T11:00:00Z"),
				EndDateTime:   mustMakeTime("2021-09-29T23:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A - Odd",
				StartDateTime: mustMakeTime("2021-10-02T11:00:00Z"),
				EndDateTime:   mustMakeTime("2021-10-02T23:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
			{
				ID:            "",
				IsActive:      true,
				Description:   "Shift A - Even",
				StartDateTime: mustMakeTime("2021-10-03T11:00:00Z"),
				EndDateTime:   mustMakeTime("2021-10-03T23:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
		},
		Error: false,
	},
}

func TestWorkCalendarGetEntries(t *testing.T) {
	for _, tc := range testCasesGetEntries {
		entries, err := tc.WorkCalendar.GetEntries()

		// Expect Error
		if tc.Error && err == nil {
			t.Errorf("Test Case '%s': got no error; want error", tc.Name)
		}

		// Don't Expect Error
		if !tc.Error && err != nil {
			t.Errorf("Test Case '%s': got error: %s; want no error", tc.Name, err)
		}

		// Expect Same Entry Count
		if len(entries) != len(tc.Entries) {
			t.Errorf("Test Case '%s': got %d entries; want %d", tc.Name, len(entries), len(tc.Entries))
		}

		// Expect Same Entries
		for _, expectedEntry := range tc.Entries {
			found := false
			for _, actualEntry := range entries {
				if expectedEntry.Description == actualEntry.Description && expectedEntry.StartDateTime.Equal(actualEntry.StartDateTime) && expectedEntry.EndDateTime.Equal(actualEntry.EndDateTime) && actualEntry.EntryType == expectedEntry.EntryType {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Test Case '%s': expected to find entry %v; but didn't", tc.Name, expectedEntry)
			}
		}
	}

	t.Log("Complete TestWorkCalendarGetEntries")
}
