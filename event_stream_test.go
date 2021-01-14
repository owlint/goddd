package goddd

import "testing"

func TestEventStreamCreation(t *testing.T) {
	stream := NewEventStream("objectId")

	if stream.ObjectId() != "objectId" {
		t.Log("Wrong object id")
		t.Fail()
	}

	if len(stream.Events()) > 0 {
		t.Log("Should not contain events")
		t.Fail()
	}
}

func TestAddEventToStream(t *testing.T) {
	stream := NewEventStream("objectId")

	stream.AddEvent("number_added", 3)

	if len(stream.Events()) != 1 {
		t.Log("Should contain one event")
		t.Fail()
	}

	event := stream.Events()[0]

	if event.Version() != 1 {
		t.Log("Event version should be 1")
		t.Fail()
	}

	if event.ObjectId() != "objectId" {
		t.Log("Event does not have appropriate object id")
		t.Fail()
	}

	if event.Name() != "number_added" {
		t.Log("Event does not have appropriate name")
		t.Fail()
	}

	payload := event.Payload().(int)
	if payload != 3 {
		t.Log("Bad payload")
		t.Fail()
	}
}

func TestAddMultipleEventsToStream(t *testing.T) {
	stream := NewEventStream("objectId")

	stream.AddEvent("number_added", 3)
	stream.AddEvent("number_added", 2)
	stream.AddEvent("number_added", 1)

	events := stream.Events()

	if events[0].Version() != 1 {
		t.Log("Wrong event version")
		t.Fail()
	}
	if events[1].Version() != 2 {
		t.Log("Wrong event version")
		t.Fail()
	}
	if events[2].Version() != 3 {
		t.Log("Wrong event version")
		t.Fail()
	}

	eventIds := make([]string, 3)

	for idx, event := range events {
		eventIds[idx] = event.Id()
	}

}

func allDifferent(arr []string) {

}

func contains(arr)
