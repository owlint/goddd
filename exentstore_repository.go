package goddd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type exentstoreNewEvent struct {
	EventType string `json:"event_type"`
	EventBody []byte `json:"event_body"`
}

type exentstoreAppendQuery struct {
	ExpectedVersion int                  `json:"expected_version"`
	Events          []exentstoreNewEvent `json:"events"`
}

type exentstoreStreamEvent struct {
	EventBody []byte `json:"event_body"`
}

type exentstoreStream struct {
	Events []exentstoreStreamEvent `json:"events"`
}

type ExentStoreRepository struct {
	client    http.Client
	endpoint  string
	publisher *EventPublisher
}

func NewExentStoreRepository(endpoint string, publisher *EventPublisher) ExentStoreRepository {
	return ExentStoreRepository{
		client:    http.Client{},
		endpoint:  endpoint,
		publisher: publisher,
	}
}

func (r *ExentStoreRepository) Save(object DomainObject) error {
	var objectRepoEvents []Event

	exist, err := r.Exists(object.ObjectID())
	if err != nil {
		return err
	}

	if exist {
		objectRepoEvents, err = r.objectRepositoryEvents(object.ObjectID())
		if err != nil {
			return err
		}
	} else {
		objectRepoEvents = make([]Event, 0)
	}

	objectEvents := object.Events()
	eventToAdd := unsavedEvents(objectEvents, objectRepoEvents)

	err = r.insertEvents(object.ObjectID(), eventToAdd)
	if err != nil {
		return err
	}

	r.publisher.Publish(eventToAdd)

	return nil
}

func (r *ExentStoreRepository) Load(objectID string, object DomainObject) error {
	exist, err := r.Exists(objectID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("Cannot load unknown object")
	}

	objectEvents, err := r.objectRepositoryEvents(objectID)
	if err != nil {
		return err
	}

	for _, event := range objectEvents {
		object.LoadEvent(object, event)
	}

	return nil
}

func (r *ExentStoreRepository) Exists(objectID string) (bool, error) {
	resp, err := r.client.Head(fmt.Sprintf("%s/api/streams/%s", r.endpoint, objectID))

	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200, nil
}

func (r *ExentStoreRepository) objectRepositoryEvents(objectID string) ([]Event, error) {
	resp, err := r.client.Get(fmt.Sprintf("%s/api/streams/%s?after_stream_version=0", r.endpoint, objectID))

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 422 || resp.StatusCode == 404 {
		return nil, errors.New("Unknown stream of bad request")
	}

	defer resp.Body.Close()

	stream := exentstoreStream{
		Events: make([]exentstoreStreamEvent, 0),
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &stream)
	if err != nil {
		return nil, err
	}

	return r.unmarshal(stream)
}

func (r *ExentStoreRepository) insertEvents(streamName string, events []Event) error {
	couldInsert := false
	var err error
	var actualVersion int
	for !couldInsert {
		actualVersion, err = r.tryInsertEvent(streamName, events)
		if actualVersion == 0 {
			couldInsert = true
		} else {
			events = r.correctEventsVersion(actualVersion, events)
		}
	}

	return err
}

func (r *ExentStoreRepository) tryInsertEvent(streamName string, events []Event) (int, error) {
	query, err := r.marshal(events)
	if err != nil {
		return 0, err
	}

	jsonQuery, err := json.Marshal(query)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/streams/%s", r.endpoint, streamName), bytes.NewBuffer(jsonQuery))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return 0, err
	}

	resp, err := r.client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 201 && resp.StatusCode != 422 {
		body, _ := ioutil.ReadAll(resp.Body)
		return 0, fmt.Errorf("Invalid status code %d : %s", resp.StatusCode, body)
	}

	if resp.StatusCode == 422 {
		return r.detectInvalidEventsError(resp.Body)
	}

	return 0, nil
}

func (r *ExentStoreRepository) correctEventsVersion(actualVersion int, events []Event) []Event {
	fmt.Println("correction", actualVersion)

	correctedEvents := make([]Event, len(events))
	for i, event := range events {
		correctedEvents[i] = NewEvent(
			event.Id(),
			event.Name(),
			actualVersion+1,
			event.Payload(),
		)
	}

	return correctedEvents
}

func (r *ExentStoreRepository) detectInvalidEventsError(body io.ReadCloser) (int, error) {
	type invalidVersion struct {
		ActualVersion int    `json:"actual_version,omitempty"`
		Error         string `json:"error,omitempty"`
	}

	bodyContent, err := ioutil.ReadAll(body)
	if err != nil {
		return 0, err
	}

	errorPayload := &invalidVersion{}
	err = json.Unmarshal(bodyContent, errorPayload)
	if err != nil {
		return 0, err
	}

	if errorPayload.Error != "" && errorPayload.ActualVersion > 0 {
		return errorPayload.ActualVersion, errors.New(errorPayload.Error)
	}

	return 0, errors.New("Unknown error")
}

func (r *ExentStoreRepository) unmarshal(stream exentstoreStream) ([]Event, error) {
	records := make([]record, len(stream.Events))
	for i, e := range stream.Events {
		record := record{}
		err := json.Unmarshal(e.EventBody, &record)
		if err != nil {
			return nil, err
		}
		records[i] = record
	}

	return fromRecords(records), nil
}

func (r *ExentStoreRepository) marshal(events []Event) (*exentstoreAppendQuery, error) {
	queryEvents := make([]exentstoreNewEvent, len(events))
	for i, event := range events {
		r, err := json.Marshal(record{
			ID:        event.Id(),
			Version:   event.Version(),
			ObjectID:  event.ObjectId(),
			Timestamp: event.Timestamp(),
			Name:      event.Name(),
			Payload:   event.Payload(),
		})
		if err != nil {
			return nil, err
		}
		queryEvents[i] = exentstoreNewEvent{
			EventType: event.Name(),
			EventBody: r,
		}
	}
	query := exentstoreAppendQuery{
		ExpectedVersion: events[0].Version() - 1,
		Events:          queryEvents,
	}

	return &query, nil
}
