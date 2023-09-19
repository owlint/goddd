package goddd

import (
	"context"
	"errors"
	"time"
)

type InMemoryRepository[T DomainObject] struct {
	eventStream []Event
	publisher   *EventPublisher
}

func NewInMemoryRepository[T DomainObject](publisher *EventPublisher) InMemoryRepository[T] {
	return InMemoryRepository[T]{
		eventStream: make([]Event, 0),
		publisher:   publisher,
	}
}

func (r *InMemoryRepository[T]) Save(ctx context.Context, object T) error {
	eventToAdd := object.CollectUnsavedEvents()

	r.eventStream = append(r.eventStream, eventToAdd...)

	r.publisher.Publish(eventToAdd)

	return nil
}

func (r *InMemoryRepository[T]) Load(ctx context.Context, objectID string, object T) error {
	if exist, err := r.Exists(ctx, objectID); err != nil || !exist {
		return errors.New("Cannot load unknown object")
	}

	object.Clear()

	objectEvents := r.objectRepositoryEvents(objectID)
	for _, event := range objectEvents {
		object.LoadEvent(object, event)
	}

	return nil
}

func (r *InMemoryRepository[T]) Exists(ctx context.Context, objectId string) (bool, error) {
	for _, event := range r.eventStream {
		if event.ObjectId() == objectId && event.name != REMOVED_EVENT_NAME {
			return true, nil
		}
	}
	return false, nil
}
func (r *InMemoryRepository[T]) objectRepositoryEvents(objectId string) []Event {
	events := make([]Event, 0)

	for _, event := range r.eventStream {
		if event.ObjectId() == objectId {
			events = append(events, event)
		}
	}

	return events
}

func (r *InMemoryRepository[T]) EventsSince(ctx context.Context, timestamp time.Time, limit int) ([]Event, error) {
	events := make([]Event, 0)

	for _, event := range r.eventStream {
		if event.Timestamp() >= timestamp.UnixNano() && len(events) < limit {
			events = append(events, event)
		}
	}

	return events, nil
}

func (r *InMemoryRepository[T]) Update(ctx context.Context, objectID string, object T, nbRetries int, updater func(T) (T, error)) (T, error) {
	return repoUpdate[T](ctx, r, objectID, object, nbRetries, updater)
}

func (r *InMemoryRepository[T]) Remove(ctx context.Context, objectID string, object T) error {
	if exists, err := r.Exists(ctx, objectID); err != nil || !exists {
		return errors.New("unknown object")
	}

	eventsToKeep := make([]Event, 0)
	for _, event := range r.eventStream {
		if event.objectID != objectID {
			eventsToKeep = append(eventsToKeep, event)
		}
	}
	r.eventStream = eventsToKeep

	event := NewEvent(objectID, REMOVED_EVENT_NAME, object.LastVersion(), []byte{})
	r.eventStream = append(r.eventStream, event)

	return nil
}
