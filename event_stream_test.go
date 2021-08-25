package goddd

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestStreamCreation(t *testing.T) {
	stream := NewEventStream()
	if len(stream.Events()) > 0 {
		t.Log("Should not contain events")
		t.Fail()
	}
}

func TestAddEventToStream(t *testing.T) {
	object := &Student{}
	stream := object.Stream

	payload := GradeSet{"b"}
	stream.AddEvent(object, "GradeSet", payload)

	assert.Len(t, stream.Events(), 1)

	event := stream.Events()[0]

	assert.Equal(t, 0, event.Version())
    assert.Equal(t, "ObjectID", event.ObjectId())
	assert.Equal(t, "GradeSet", event.Name())

	assert.Equal(t, "b", object.grade)
}

func TestAddMultipleEventsToStream(t *testing.T) {
	object := &Student{}
	stream := object.Stream

	payload := GradeSet{}
	payload.Grade = "3"
	stream.AddEvent(object, "GradeSet", payload)
	payload.Grade = "2"
	stream.AddEvent(object, "GradeSet", payload)
	payload.Grade = "1"
	stream.AddEvent(object, "GradeSet", payload)

	events := stream.Events()

	assert.Equal(t, 0, events[0].Version())
	assert.Equal(t, 1, events[1].Version())
	assert.Equal(t, 2, events[2].Version())
	assert.Equal(t, 2, events[2].Version())

	eventIds := make([]string, 3)
	for idx, event := range events {
		eventIds[idx] = event.Id()
	}

	assert.True(t, allDifferent(eventIds))
}

func TestLastVersion(t *testing.T) {
	object := &Student{}
	stream := object.Stream

	payload := GradeSet{}
	payload.Grade = "3"
	stream.AddEvent(object, "GradeSet", payload)
	payload.Grade = "2"
	stream.AddEvent(object, "GradeSet", payload)
	payload.Grade = "1"
	stream.AddEvent(object, "GradeSet", payload)

	assert.Equal(t, 3, stream.LastVersion())
}

func TestContainsEvent(t *testing.T) {
	object := &Student{}
	stream := object.Stream

	payload := GradeSet{}
	payload.Grade = "3"
	stream.AddEvent(object, "GradeSet", payload)
	payload.Grade = "2"
	stream.AddEvent(object, "GradeSet", payload)
	payload.Grade = "1"

	assert.True(t, stream.ContainsEventWithId(stream.Events()[0].Id()))
}

func TestNotContainEvent(t *testing.T) {
	object := &Student{}
	stream := object.Stream

	payload := GradeSet{}
	payload.Grade = "3"
	stream.AddEvent(object, "GradeSet", payload)
	payload.Grade = "2"
	stream.AddEvent(object, "GradeSet", payload)
	payload.Grade = "1"

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
