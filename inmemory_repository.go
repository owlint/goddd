package goddd

import (
	"errors"
	"time"
)

type InMemoryRepository struct {
	eventStream []Event
	publisher   *EventPublisher
}

func NewInMemoryRepository(publisher *EventPublisher) InMemoryRepository {
	return InMemoryRepository{
		eventStream: make([]Event, 0),
		publisher:   publisher,
	}
}

func (r *InMemoryRepository) Save(object DomainObject) error {
	eventToAdd := object.CollectUnsavedEvents()

	r.eventStream = append(r.eventStream, eventToAdd...)

	r.publisher.Publish(eventToAdd)

	return nil
}

func (r *InMemoryRepository) Load(objectID string, object DomainObject) error {
	if exist, err := r.Exists(objectID); err != nil || !exist {
		return errors.New("Cannot load unknown object")
	}

	object.Clear()

	objectEvents := r.objectRepositoryEvents(objectID)
	for _, event := range objectEvents {
		object.LoadEvent(object, event)
	}

	return nil
}

func (r *InMemoryRepository) Exists(objectId string) (bool, error) {
	for _, event := range r.eventStream {
		if event.ObjectId() == objectId {
			return true, nil
		}
	}
	return false, nil
}
func (r *InMemoryRepository) objectRepositoryEvents(objectId string) []Event {
	events := make([]Event, 0)

	for _, event := range r.eventStream {
		if event.ObjectId() == objectId {
			events = append(events, event)
		}
	}

	return events
}

func (r *InMemoryRepository) EventsSince(timestamp time.Time, limit int) ([]Event, error) {
	events := make([]Event, 0)

	for _, event := range r.eventStream {
		if event.Timestamp() >= timestamp.UnixNano() && len(events) < limit {
			events = append(events, event)
		}
	}

	return events, nil
}
