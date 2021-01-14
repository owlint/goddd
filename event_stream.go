package goddd

type EventStream interface {
	ObjectId() string
	AddEvent(eventName string, payload interface{})
	Events() []Event
}

type eventStream struct {
	objectId    string
	events      []Event
	lastVersion int
}

func (stream eventStream) ObjectId() string {
	return stream.objectId
}

func (stream *eventStream) AddEvent(eventName string, payload interface{}) {
	event := NewEvent(stream.objectId, eventName, stream.lastVersion+1, payload)
	stream.events = append(stream.events, event)
	stream.lastVersion += 1
}

func (stream eventStream) Events() []Event {
	return stream.events
}

func NewEventStream(objectId string) EventStream {
	var stream EventStream
	stream = &eventStream{
		objectId:    objectId,
		events:      make([]Event, 0),
		lastVersion: 0,
	}

	return stream
}
