package goddd

import (
	"fmt"

	"github.com/google/uuid"
)

type Repository interface {
	Save(object DomainObject) error
	Load(object DomainObject) error
	Exists(objectID string) (bool, error)
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
