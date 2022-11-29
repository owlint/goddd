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

func unsavedEvents(objectEvents []Event, knownEventIDs []string) []Event {
	knownIDs := make(map[string]struct{}, len(knownEventIDs))
	events := make([]Event, 0)

	for _, eventID := range knownEventIDs {
		knownIDs[eventID] = struct{}{}
	}

	for _, event := range objectEvents {
		if _, exists := knownIDs[event.Id()]; !exists {
			events = append(events, event)
		}
	}

	return events
}

func NewIdentity(objectType string) string {
	return fmt.Sprintf("%s-%s", objectType, uuid.New().String())
}

func Encode(object interface{}) ([]byte, error) {
	var data bytes.Buffer
	encoder := gob.NewEncoder(&data)
	err := encoder.Encode(object)
	if err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}

func Decode(object interface{}, data []byte) error {
	encoder := gob.NewDecoder(bytes.NewBuffer(data))
	err := encoder.Decode(object)
	if err != nil {
		return err
	}
	return nil
}
