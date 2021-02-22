package goddd

import "errors"

type InMemoryRepository struct {
	eventStream []Event
}

func NewInMemoryRepository() InMemoryRepository {
	return InMemoryRepository{
		eventStream: make([]Event, 0),
	}
}

func (r *InMemoryRepository) Save(object DomainObject) error {
	objectRepoEvents := r.objectRepositoryEvents(object.ObjectID())
	objectEvents := object.Events()
	eventToAdd := unsavedEvents(objectEvents, objectRepoEvents)

	for _, event := range eventToAdd {
		r.eventStream = append(r.eventStream, event)
	}

	return nil
}

func (r *InMemoryRepository) Load(object DomainObject) error {
	objectID := object.ObjectID()

	if exist, err := r.Exists(objectID); err != nil || !exist {
		return errors.New("Cannot load unknown object")
	}

	objectEvents := r.objectRepositoryEvents(objectID)
	for _, event := range objectEvents {
		object.AddEvent(object, event.Name(), event.Payload())
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
