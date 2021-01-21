package goddd

type EventStream interface {
	AddEvent(eventName string, payload interface{})
	Events() []Event
	LastVersion() int
}

type eventStream struct {
	object      DomainObject
	events      []Event
	lastVersion int
}

func (stream *eventStream) AddEvent(eventName string, payload interface{}) {
	event := NewEvent(stream.object.ObjectId(), eventName, stream.lastVersion+1, payload)
	stream.events = append(stream.events, event)
	stream.lastVersion += 1
	stream.object.Apply(eventName, payload)
}

func (stream eventStream) Events() []Event {
	return stream.events
}

func (stream eventStream) LastVersion() int {
	return stream.lastVersion
}

func NewEventStream(object DomainObject) EventStream {
	var stream EventStream
	stream = &eventStream{
		object:      object,
		events:      make([]Event, 0),
		lastVersion: 0,
	}

	return stream
}

func ReloadEventStream(object DomainObject, events []Event) EventStream {
	maxVersion := 0

	for _, event := range events {
		object.Apply(event.Name(), event.Payload())
		if event.Version() > maxVersion {
			maxVersion = event.Version()
		}
	}

	var stream EventStream
	stream = &eventStream{
		object:      object,
		events:      events,
		lastVersion: maxVersion,
	}

	return stream
}
