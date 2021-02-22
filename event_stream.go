package goddd

type EventStream interface {
	AddEvent(object DomainObject, eventName string, payload interface{})
	Events() []Event
	LastVersion() int
	ContainsEventWithId(eventID string) bool
}

type Stream struct {
	events      []Event
	lastVersion int
}

func (s *Stream) AddEvent(object DomainObject, eventName string, payload interface{}) {
	event := NewEvent(object.ObjectID(), eventName, s.lastVersion+1, payload)
	s.events = append(s.events, event)
	s.lastVersion++
	object.Apply(eventName, payload)
}

func (s *Stream) Events() []Event {
	return s.events
}

func (s *Stream) LastVersion() int {
	return s.lastVersion
}

func (s *Stream) ContainsEventWithId(eventId string) bool {
	for _, event := range s.Events() {
		if event.Id() == eventId {
			return true
		}
	}
	return false
}

func NewEventStream() Stream {
	return Stream{
		events:      make([]Event, 0),
		lastVersion: 0,
	}
}
