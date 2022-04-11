package goddd

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Save(object DomainObject) error
	Load(objectID string, object DomainObject) error
	Exists(objectID string) (bool, error)
	EventsSince(time time.Time, limit int) ([]Event, error)
}

func unsavedEvents(objectEvents []Event, alreadySavedObjectEvents []Event) []Event {
	events := make([]Event, 0)

	for _, event := range objectEvents {
		if !isEventAlreadySaved(event, alreadySavedObjectEvents) {
			events = append(events, event)
		}
	}

	return events
}

func isEventAlreadySaved(event Event, knownEvents []Event) bool {
	for _, knownEvent := range knownEvents {
		if knownEvent.Id() == event.Id() {
			return true
		}
	}
	return false
}

func NewIdentity(objectType string) string {
	return fmt.Sprintf("%s-%s", objectType, uuid.New().String())
}

func Encode(object DomainObject) ([]byte, error) {
	var data bytes.Buffer
	encoder := gob.NewEncoder(&data)
	err := encoder.Encode(object)
	if err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}

func Decode(data []byte) (*DomainObject, error) {
	var object *DomainObject
	encoder := gob.NewDecoder(bytes.NewBuffer(data))
	err := encoder.Decode(object)
	if err != nil {
		return nil, err
	}

	return object, nil
}
