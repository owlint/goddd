package goddd

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	id        string
	version   int
	objectId  string
	timestamp int64
	name      string
	payload   interface{}
}

func (event Event) Id() string {
	return event.id
}

func (event Event) Version() int {
	return event.version
}

func (event Event) ObjectId() string {
	return event.objectId
}

func (event Event) Timestamp() int64 {
	return event.timestamp
}

func (event Event) Name() string {
	return event.name
}

func (event Event) Payload() interface{} {
	return event.payload
}

func NewEvent(objectId string, eventName string, version int, payload interface{}) Event {
	return Event{
		id:        fmt.Sprintf("%s-%s", objectId, uuid.New().String()),
		version:   version,
		objectId:  objectId,
		timestamp: time.Now().UnixNano(),
		name:      eventName,
		payload:   payload,
	}
}
