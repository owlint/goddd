package goddd

import "github.com/tinylib/msgp/msgp"

// DomainObject is an interface representing a domain object necessary methods
type DomainObject interface {
	EventStream

	ObjectID() string
	Apply(eventName string, eventPayload []byte) error
}

// DomainObjectMemento is an interface representing a domain object capable of exposing a memento
type DomainObjectMemento interface {
	DumpMemento() (msgp.Marshaler, error)
	ApplyMemento(payload []byte) error
	SetVersion(version int)
}
