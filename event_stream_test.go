package goddd

import (
	"testing"

	"github.com/google/uuid"
	"github.com/owlint/goddd/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestEventStreamCreation(t *testing.T) {
	stream := NewEventStream()
	if len(stream.Events()) > 0 {
		t.Log("Should not contain events")
		t.Fail()
	}
}

func TestAddEventToStream(t *testing.T) {
	object := testDomainObject{}
	stream := object.Stream

	payload := &pb.NumberAdded{}
	payload.Nb = 3
	stream.AddEvent(&object, "NumberAdded", payload)

	assert.Len(t, stream.Events(), 1)

	event := stream.Events()[0]

	assert.Equal(t, 0, event.Version())
	assert.Equal(t, "objectId", event.ObjectId())
	assert.Equal(t, "NumberAdded", event.Name())

	payload.Reset()
	err := proto.Unmarshal(event.Payload(), payload)
	assert.Nil(t, err)

	assert.Equal(t, int32(3), payload.Nb)
	assert.Equal(t, int32(3), object.number)
}

func TestAddMultipleEventsToStream(t *testing.T) {
	object := testDomainObject{}
	stream := object.Stream

	payload := &pb.NumberAdded{}
	payload.Nb = 3
	stream.AddEvent(&object, "number_added", payload)
	payload.Reset()
	payload.Nb = 2
	stream.AddEvent(&object, "number_added", payload)
	payload.Reset()
	payload.Nb = 1
	stream.AddEvent(&object, "number_added", payload)

	events := stream.Events()

	assert.Equal(t, 0, events[0].Version())
	assert.Equal(t, 1, events[1].Version())
	assert.Equal(t, 2, events[2].Version())

	eventIds := make([]string, 3)
	for idx, event := range events {
		eventIds[idx] = event.Id()
	}

	assert.True(t, allDifferent(eventIds))
}

func TestLastVersion(t *testing.T) {
	object := testDomainObject{}
	stream := object.Stream

	payload := &pb.NumberAdded{}
	payload.Nb = 3
	stream.AddEvent(&object, "number_added", payload)
	payload.Reset()
	payload.Nb = 2
	stream.AddEvent(&object, "number_added", payload)
	payload.Reset()
	payload.Nb = 1
	stream.AddEvent(&object, "number_added", payload)

	assert.Equal(t, 3, stream.LastVersion())
}

func TestContainsEvent(t *testing.T) {
	object := testDomainObject{}
	stream := object.Stream

	payload := &pb.NumberAdded{}
	payload.Nb = 3
	stream.AddEvent(&object, "number_added", payload)
	payload.Reset()
	payload.Nb = 2
	stream.AddEvent(&object, "number_added", payload)
	payload.Reset()
	payload.Nb = 1

	assert.True(t, stream.ContainsEventWithId(stream.Events()[0].Id()))
}

func TestNotContainEvent(t *testing.T) {
	object := testDomainObject{}
	stream := object.Stream

	payload := &pb.NumberAdded{}
	payload.Nb = 3
	stream.AddEvent(&object, "number_added", payload)
	payload.Reset()
	payload.Nb = 2
	stream.AddEvent(&object, "number_added", payload)
	payload.Reset()
	payload.Nb = 1

	assert.False(t, stream.ContainsEventWithId(uuid.New().String()))
}

func allDifferent(arr []string) bool {
	return len(arrayToSet(arr)) == len(arr)
}

func arrayToSet(arr []string) []string {
	set := make([]string, 0)

	for _, element := range arr {
		if !contains(set, element) {
			set = append(set, element)
		}
	}

	return set
}

func contains(arr []string, element string) bool {
	for _, str := range arr {
		if str == element {
			return true
		}
	}
	return false
}
