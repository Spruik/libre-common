package services

import (
	"errors"
	"testing"
	"time"

	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
)

var stdMessageChan chan domain.StdMessageStruct

type FakeLibreConnector struct {
	nextError bool
}

func (libreConnector FakeLibreConnector) Connect() error {
	return nil
}

func (libreConnector FakeLibreConnector) Close() error {
	return nil
}

func (libreConnector FakeLibreConnector) SendStdMessage(msg domain.StdMessageStruct) error {

	go func() { stdMessageChan <- msg }()

	if libreConnector.nextError {
		return errors.New("Some Error")
	}
	return nil
}

func (libreConnector FakeLibreConnector) SetNextError() {
}

func (libreConnector FakeLibreConnector) ListenForReadTagsRequest(c chan []domain.StdMessageStruct, readTagDefs []domain.StdMessageStruct) {
}

func (libreConnector FakeLibreConnector) ListenForWriteTagsRequest(c chan []domain.StdMessageStruct, writeTagDefs []domain.StdMessageStruct) {
}

func (libreConnector FakeLibreConnector) ListenForGetTagHistoryRequest(c chan []domain.StdMessageStruct, startTS time.Time, endTS time.Time, inTagDefs []domain.StdMessageStruct) {
}

type FakeLibreDataStore struct {
	WorkCalendars []domain.WorkCalendar
	IsNextError   bool
	Err           error
}

func (fakeLibreDataStore FakeLibreDataStore) SetNextError(err error) {
}

func (fakeLibreDataStore FakeLibreDataStore) Connect() error {
	return nil
}

//Close is called to close the data store connection
func (fakeLibreDataStore FakeLibreDataStore) Close() error {
	return nil
}

//BeginTransaction starts a transaction with the data store and returns a handle for use with operations
func (fakeLibreDataStore FakeLibreDataStore) BeginTransaction(forUpdate bool, name string) ports.LibreDataStoreTransactionPort {
	return nil
}

//GetSubscription returns a handle to a database subscription
func (fakeLibreDataStore FakeLibreDataStore) GetSubscription(q interface{}, vars map[string]interface{}) ports.LibreDataStoreSubscriptionPort {
	return nil
}

//GetAllActiveWorkCalendar returns the work calendars
func (fakeLibreDataStore FakeLibreDataStore) GetAllActiveWorkCalendar() ([]domain.WorkCalendar, error) {
	if fakeLibreDataStore.IsNextError {
		return fakeLibreDataStore.WorkCalendars, fakeLibreDataStore.Err
	}
	return fakeLibreDataStore.WorkCalendars, nil
}

func TestCalendarService(t *testing.T) {
	stdMessageChan = make(chan domain.StdMessageStruct)

	now := time.Now().UTC()
	fakeLibreConnector := FakeLibreConnector{}
	fakeLibreDataStore := FakeLibreDataStore{
		WorkCalendars: []domain.WorkCalendar{
			{
				ID:          "",
				Name:        "",
				IsActive:    true,
				Description: "",
				Definition: []domain.WorkCalendarDefinitionEntry{
					{
						ID:            "",
						IsActive:      true,
						Description:   "Shift A",
						Freq:          domain.Daily,
						StartDateTime: now.Add(-5 * 24 * time.Hour),
						EndDateTime:   now.Add(5 * 24 * time.Hour),
						Count:         31,
						Interval:      1,
						Weekday:       domain.Sunday,
						Duration:      "PT8H",
						EntryType:     domain.PlannedBusyTime,
					},
				},
				Equipment: []domain.Equipment{
					{
						Id:          "",
						Name:        "Site/Area/Line",
						Description: "",
					},
				},
				Entries: []domain.WorkCalendarEntry{
					{
						ID:            "",
						IsActive:      true,
						Description:   "Shift A",
						StartDateTime: now.Add(-5 * 24 * time.Hour),
						EndDateTime:   now.Add(-5 * 24 * time.Hour).Add(8 * time.Hour),
						EntryType:     domain.PlannedBusyTime,
					},
					{
						ID:            "",
						IsActive:      true,
						Description:   "Shift A",
						StartDateTime: now.Add(-4 * 24 * time.Hour),
						EndDateTime:   now.Add(-4 * 24 * time.Hour).Add(8 * time.Hour),
						EntryType:     domain.PlannedBusyTime,
					},
					{
						ID:            "",
						IsActive:      true,
						Description:   "Shift A",
						StartDateTime: now.Add(-3 * 24 * time.Hour),
						EndDateTime:   now.Add(-3 * 24 * time.Hour).Add(8 * time.Hour),
						EntryType:     domain.PlannedBusyTime,
					},
					{
						ID:            "",
						IsActive:      true,
						Description:   "Shift A",
						StartDateTime: now.Add(-2 * 24 * time.Hour),
						EndDateTime:   now.Add(-2 * 24 * time.Hour).Add(8 * time.Hour),
						EntryType:     domain.PlannedBusyTime,
					},
					{
						ID:            "",
						IsActive:      true,
						Description:   "Shift A",
						StartDateTime: now.Add(-1 * 24 * time.Hour),
						EndDateTime:   now.Add(-1 * 24 * time.Hour).Add(8 * time.Hour),
						EntryType:     domain.PlannedBusyTime,
					},
					{
						ID:            "",
						IsActive:      true,
						Description:   "Shift A",
						StartDateTime: now.Add(time.Hour),
						EndDateTime:   now.Add(8 * time.Hour),
						EntryType:     domain.PlannedBusyTime,
					},
					{
						ID:            "",
						IsActive:      true,
						Description:   "Shift A",
						StartDateTime: now.Add(1 * 24 * time.Hour),
						EndDateTime:   now.Add(1 * 24 * time.Hour).Add(8 * time.Hour),
						EntryType:     domain.PlannedBusyTime,
					},
					{
						ID:            "",
						IsActive:      true,
						Description:   "Shift A",
						StartDateTime: now.Add(2 * 24 * time.Hour),
						EndDateTime:   now.Add(2 * 24 * time.Hour).Add(8 * time.Hour),
						EntryType:     domain.PlannedBusyTime,
					},
					{
						ID:            "",
						IsActive:      true,
						Description:   "Shift A",
						StartDateTime: now.Add(3 * 24 * time.Hour),
						EndDateTime:   now.Add(3 * 24 * time.Hour).Add(8 * time.Hour),
						EntryType:     domain.PlannedBusyTime,
					},
					{
						ID:            "",
						IsActive:      true,
						Description:   "Shift A",
						StartDateTime: now.Add(4 * 24 * time.Hour),
						EndDateTime:   now.Add(4 * 24 * time.Hour).Add(8 * time.Hour),
						EntryType:     domain.PlannedBusyTime,
					},
					{
						ID:            "",
						IsActive:      true,
						Description:   "Shift A",
						StartDateTime: now.Add(5 * 24 * time.Hour),
						EndDateTime:   now.Add(5 * 24 * time.Hour).Add(8 * time.Hour),
						EntryType:     domain.PlannedBusyTime,
					},
				},
			},
		},
	}

	libreConfig.Initialize("../../../config/calender-test-config.json")
	libreLogger.Initialize("libreLogger")

	service := NewCalendarService("calendarService", fakeLibreDataStore, fakeLibreConnector)

	workCalendars, err := service.GetAllActiveWorkCalendar()

	if err != nil {
		t.Errorf("TestCalendarService failed expected no error; got %s", err)
	}

	if len(workCalendars) != len(fakeLibreDataStore.WorkCalendars) {
		t.Errorf("TestCalendarService failed expected %d entries; got %d", len(fakeLibreDataStore.WorkCalendars), len(workCalendars))
	}

	for i := range workCalendars {
		if workCalendars[i].ID != fakeLibreDataStore.WorkCalendars[i].ID ||
			workCalendars[i].Name != fakeLibreDataStore.WorkCalendars[i].Name ||
			workCalendars[i].Description != fakeLibreDataStore.WorkCalendars[i].Description ||
			len(workCalendars[i].Definition) != len(fakeLibreDataStore.WorkCalendars[i].Definition) ||
			len(workCalendars[i].Entries) != len(fakeLibreDataStore.WorkCalendars[i].Entries) ||
			len(workCalendars[i].Equipment) != len(fakeLibreDataStore.WorkCalendars[i].Equipment) {
			t.Errorf("TestCalendarService failed comparing WorkCalenders at index %d, expected %v; got %v", i, workCalendars[i], fakeLibreDataStore.WorkCalendars[i])
		}

		// Compare Definition
		for j := range workCalendars[i].Definition {
			if workCalendars[i].Definition[j].ID != fakeLibreDataStore.WorkCalendars[i].Definition[j].ID {
				t.Errorf("TestCalendarService failed comparing WorkCalenders Definitions at index %d:%d, expected %v; got %v", i, j, workCalendars[i].Definition[j], fakeLibreDataStore.WorkCalendars[i].Definition[j])
			}
		}

		// Compare Equipment
		for j := range workCalendars[i].Equipment {
			if workCalendars[i].Equipment[j].Id != fakeLibreDataStore.WorkCalendars[i].Equipment[j].Id ||
				workCalendars[i].Equipment[j].Name != fakeLibreDataStore.WorkCalendars[i].Equipment[j].Name {
				t.Errorf("TestCalendarService failed comparing WorkCalenders Equipment at index %d:%d, expected %v; got %v", i, j, workCalendars[i].Equipment[j], fakeLibreDataStore.WorkCalendars[i].Equipment[j])
			}
		}

		// Compare Entries
		for j := range workCalendars[i].Entries {
			if workCalendars[i].Entries[j] != fakeLibreDataStore.WorkCalendars[i].Entries[j] {
				t.Errorf("TestCalendarService failed comparing WorkCalenders Entries at index %d:%d, expected %v; got %v", i, j, workCalendars[i].Entries[j], fakeLibreDataStore.WorkCalendars[i].Entries[j])
			}
		}
	}
	service.SetTickSpeed(time.Second * 1)
	service.Start()
	time.Sleep(20 * time.Millisecond)

	expectedMessage := domain.StdMessageStruct{
		OwningAsset:      fakeLibreDataStore.WorkCalendars[0].Equipment[0].Name,
		OwningAssetId:    fakeLibreDataStore.WorkCalendars[0].Equipment[0].Id,
		ItemName:         "workCalendarCategory",
		ItemNameExt:      map[string]string{},
		ItemId:           "",
		ItemValue:        string(domain.PlannedBusyTime),
		ItemDataType:     "STRING",
		TagQuality:       1,
		Err:              nil,
		ChangedTimestamp: time.Now().UTC(),
		Category:         "TAGDATA",
		Topic:            fakeLibreDataStore.WorkCalendars[0].Equipment[0].Name + "/workCalendarCategory",
	}

	actualMessage := <-stdMessageChan

	if expectedMessage.OwningAsset != actualMessage.OwningAsset ||
		expectedMessage.OwningAssetId != actualMessage.OwningAssetId ||
		expectedMessage.ItemName != actualMessage.ItemName ||
		expectedMessage.ItemId != actualMessage.ItemId ||
		expectedMessage.ItemValue != actualMessage.ItemValue ||
		expectedMessage.ItemDataType != actualMessage.ItemDataType ||
		expectedMessage.TagQuality != actualMessage.TagQuality ||
		expectedMessage.Err != actualMessage.Err ||
		expectedMessage.Category != actualMessage.Category ||
		expectedMessage.Topic != actualMessage.Topic {
		t.Errorf("TestCalendarService failed comparing StdMessageStructs want %v; got %v", expectedMessage, actualMessage)
	}
	// Start it again
	service.Start()
	time.Sleep(20 * time.Millisecond)
	service.Stop()
	time.Sleep(1 * time.Second)

	SetCalendarServiceInstance(service)
	if GetCalendarServiceInstance() != service {
		t.Errorf("TestCalendarService get/set serivceInstance failed")
	}

	t.Logf("Complete CalendarService")
}
