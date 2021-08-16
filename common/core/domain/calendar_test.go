package domain

import (
	"testing"
	"time"
)

func mustMakeTime(str string) (t time.Time) {
	t, _ = time.Parse(time.RFC3339, str)
	return t
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
			Id:          "abc123",
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
			Id:          "abc123",
			Name:        "Work Calendar - Single Definition",
			Description: "Description",
			Entries:     []WorkCalendarEntry{},
			Equipment:   []Equipment{},
			Definition: []WorkCalendarDefinitionEntry{
				{
					Id:            "abc123",
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
				Id:            "abc123",
				IsActive:      true,
				Description:   "Shift A",
				StartDateTime: mustMakeTime("2021-08-13T08:00:00Z"),
				EndDateTime:   mustMakeTime("2021-08-13T16:00:00Z"),
				EntryType:     PlannedBusyTime,
			},
		},
		Error: false,
	},
}

func TestWorkCalendarGetEntries(t *testing.T) {
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
				if expectedEntry.Description == actualEntry.Description && expectedEntry.StartDateTime == actualEntry.StartDateTime && expectedEntry.EndDateTime == actualEntry.EndDateTime && actualEntry.EntryType == expectedEntry.EntryType {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Test Case '%s': expected to find entry %v; but didn't", tc.Name, expectedEntry)
			}
		}
	}

	t.Log("Complete")
}
