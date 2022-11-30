package goddd

import (
	"github.com/tinylib/msgp/msgp"
)

// EventStream is an interface representing a stream of events
type EventStream interface {
	AddEvent(object DomainObject, eventName string, payload msgp.Marshaler) error
	LoadEvent(object DomainObject, event Event) error
	Events() []Event
	CollectUnsavedEvents() []Event
	LastVersion() int
	SetStreamVersion(version int)
	ContainsEventWithId(eventID string) bool
	Clear()
}

// Stream is an implementation of an EventStream
type Stream struct {
	events        []Event
	unsavedEvents []*Event
	lastVersion   int
}

// AddEvent add a new event into the stream
func (s *Stream) AddEvent(object DomainObject, eventName string, payload msgp.Marshaler) error {
	bytePayload, err := payload.MarshalMsg(nil)
	if err != nil {
		return err
	}

	event := NewEvent(object.ObjectID(), eventName, s.lastVersion, bytePayload)
	s.unsavedEvents = append(s.unsavedEvents, &event)
	return s.LoadEvent(object, event)
}

// LoadEvent load an existing event into the stream
func (s *Stream) LoadEvent(object DomainObject, event Event) error {
	s.events = append(s.events, event)
	s.lastVersion++
	return object.Apply(event.Name(), event.Payload())
}

// Events returns all the events of this stream
func (s *Stream) Events() []Event {
	return s.events
}

func (s *Stream) CollectUnsavedEvents() []Event {
	events := make([]Event, len(s.unsavedEvents))
	for i, event := range s.unsavedEvents {
		events[i] = *event
	}
	s.unsavedEvents = make([]*Event, 0)
	return events
}

// LastVersion returns the last known version
func (s *Stream) LastVersion() int {
	return s.lastVersion
}

func (s *Stream) SetStreamVersion(version int) {
	s.lastVersion = version + 1
}

// ContainsEventWithId checks if an event is known in the stream
func (s *Stream) ContainsEventWithId(eventId string) bool {
	for _, event := range s.Events() {
		if event.Id() == eventId {
			return true
		}
	}
	return false
}

// Clear clears the stream
func (s *Stream) Clear() {
	s.events = make([]Event, 0)
	s.lastVersion = 0
}

// NewEventStream initializes a new event stream
func NewEventStream() Stream {
	return Stream{
		events:      make([]Event, 0),
		lastVersion: 0,
	}
}
