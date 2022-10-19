package goddd

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/owlint/goddd/protobuf"
	"google.golang.org/protobuf/proto"
)

// Event represents a domain Event
type Event struct {
	id        string
	version   int
	objectID  string
	timestamp int64
	name      string
	payload   []byte
}

// Id of the domain event
func (event Event) Id() string {
	return event.id
}

// Version of the domain event
func (event Event) Version() int {
	return event.version
}

// ObjectId the event is linked to
func (event Event) ObjectId() string {
	return event.objectID
}

// Timestamp of the event
func (event Event) Timestamp() int64 {
	return event.timestamp
}

// Name of the event
func (event Event) Name() string {
	return event.name
}

// Payload of the event as byte array
func (event Event) Payload() []byte {
	return event.payload
}

func (event Event) Serialize() ([]byte, error) {
	unixTimestamp := time.Unix(0, event.Timestamp()).Unix()

	pbEvent := &protobuf.EventSourcingEvent{
		Timestamp: unixTimestamp,
		ObjectID:  event.ObjectId(),
		Name:      event.Name(),
		Payload:   string(event.Payload()),
		Version:   int32(event.Version()),
	}

	return proto.Marshal(pbEvent)
}

// NewEvent create a new event from the given parameters
func NewEvent(objectID string, eventName string, version int, payload []byte) Event {
	return Event{
		id:        fmt.Sprintf("%s-%s", objectID, uuid.New().String()),
		version:   version,
		objectID:  objectID,
		timestamp: time.Now().UnixNano(),
		name:      eventName,
		payload:   payload,
	}
}

func ReloadEvent(eventID, objectID, eventName string, version int, payload []byte, timestamp int64) Event {
	return Event{
		id:        eventID,
		version:   version,
		objectID:  objectID,
		timestamp: timestamp,
		name:      eventName,
		payload:   payload,
	}
}

func Deserialize(message []byte) (Event, error) {
	pbEvent := protobuf.EventSourcingEvent{}
	proto.Unmarshal(message, &pbEvent)

	event := Event{
		timestamp: time.Unix(int64(pbEvent.Timestamp), 0).UnixNano(),
		objectID:  pbEvent.GetObjectID(),
		name:      pbEvent.GetName(),
		payload:   []byte(pbEvent.GetPayload()),
		version:   int(pbEvent.GetVersion()),
	}
	return event, nil
}
