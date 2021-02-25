package goddd

// DomainObject is an interface representing a domain object necessary methods
type DomainObject interface {
	EventStream

	ObjectID() string
	Apply(eventName string, eventPayload []byte) error
}
