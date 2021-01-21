package goddd

type DomainObject interface {
	ObjectId() string
	EventStream() EventStream
	Apply(eventName string, eventPayload interface{})
}
