package goddd

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEventCreation(t *testing.T) {
	objectId := uuid.New().String()
	before := time.Now().UnixNano()
	time.Sleep(5 * time.Millisecond)
	event := NewEvent(objectId, "eventCreated", 3, true)
	time.Sleep(5 * time.Millisecond)

	if event.Name() != "eventCreated" {
		t.Log("Wrong event name")
		t.Fail()
	}

	if event.Version() != 3 {
		t.Log("Wrong event version")
		t.Fail()
	}

	if !strings.HasPrefix(event.Id(), objectId) {
		t.Log("Event id does not contain object id")
		t.Fail()
	}

	if len(event.Id()) <= len(objectId) {
		t.Log("Event id is not bigger then object Id")
		t.Fail()
	}

	now := time.Now().UnixNano()
	if event.Timestamp() <= before || event.Timestamp() >= now {
		t.Logf("Wrong timestamp %d <= %d <= %d", before, event.Timestamp(), now)
		t.Fail()
	}

	if event.ObjectId() != objectId {
		t.Log("Wrong objectId")
		t.Fail()
	}

	castedPayload := event.Payload().(bool)
	if !castedPayload {
		t.Log("Wrong payload")
		t.Fail()
	}
}

func benchmarkEventCreation(objectId string, eventName string, version int, payload interface{}, b *testing.B) {
	for n := 0; n < b.N; n++ {
		NewEvent(objectId, eventName, version, payload)
	}
}

func BenchmarkEventCreationBool(b *testing.B) {
	objectId := uuid.New().String()
	benchmarkEventCreation(objectId, "eventName", 3, true, b)
}

func BenchmarkEventCreationArray(b *testing.B) {
	objectId := uuid.New().String()
	benchmarkEventCreation(objectId, "eventName", 3, []int{1, 2, 3, 4}, b)
}
