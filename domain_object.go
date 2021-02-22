package goddd

type DomainObject interface {
	EventStream

	ObjectID() string
	Apply(eventName string, eventPayload interface{})
}
