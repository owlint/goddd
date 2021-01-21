package goddd

import "testing"

type testDomainObject struct {
	number int
}

func (object testDomainObject) ObjectId() string {
	return "objectId"
}

func (object *testDomainObject) Apply(eventName string, event interface{}) {
	switch eventName {
	case "NumberAdded":
		toAdd := event.(int)
		object.number += toAdd
	}
}

func (o testDomainObject) EventStream() EventStream {
	return &eventStream{}
}

func TestEventStreamCreation(t *testing.T) {
	stream := NewEventStream(&testDomainObject{})

	if len(stream.Events()) > 0 {
		t.Log("Should not contain events")
		t.Fail()
	}
}

func TestAddEventToStream(t *testing.T) {
	object := testDomainObject{}
	stream := NewEventStream(&object)

	stream.AddEvent("NumberAdded", 3)

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

	if event.Name() != "NumberAdded" {
		t.Log("Event does not have appropriate name")
		t.Fail()
	}

	payload := event.Payload().(int)
	if payload != 3 {
		t.Log("Bad payload")
		t.Fail()
	}

	if object.number != 3 {
		t.Log("Event not applied")
		t.Fail()
	}
}

func TestAddMultipleEventsToStream(t *testing.T) {
	stream := NewEventStream(&testDomainObject{})

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

	if !allDifferent(eventIds) {
		t.Logf("Not all ids are different, %s", eventIds)
		t.Fail()
	}
}

func TestLastVersion(t *testing.T) {
	stream := NewEventStream(&testDomainObject{})

	stream.AddEvent("number_added", 3)
	stream.AddEvent("number_added", 2)
	stream.AddEvent("number_added", 1)

	if stream.LastVersion() != 3 {
		t.Log("Last version should be 3")
		t.Fail()
	}
}

func TestReload(t *testing.T) {
	domainObject := &testDomainObject{}
	stream := NewEventStream(domainObject)

	stream.AddEvent("number_added", 3)
	stream.AddEvent("number_added", 2)
	stream.AddEvent("number_added", 1)

	events := stream.Events()

	reloadedDomainObject := &testDomainObject{}
	reloadedStream := ReloadEventStream(reloadedDomainObject, events)

	if domainObject.number != reloadedDomainObject.number {
		t.Log("Wrong value")
		t.Fail()
	}

	if reloadedStream.LastVersion() != stream.LastVersion() {
		t.Log("Wrong last version")
		t.Fail()
	}
}

func benchmarkAddEvents(nbEvents int, payload interface{}, b *testing.B) {
	stream := NewEventStream(&testDomainObject{})
	for n := 0; n < b.N; n++ {
		for eventNb := 0; eventNb < nbEvents; eventNb++ {
			stream.AddEvent("added", payload)
		}
	}
}

func BenchmarkAddEvents100bool(b *testing.B) {
	var payload interface{}
	payload = true
	benchmarkAddEvents(100, payload, b)
}

func BenchmarkAddEvents1000bool(b *testing.B) {
	var payload interface{}
	payload = true
	benchmarkAddEvents(1000, payload, b)
}

func BenchmarkAddEvents100string(b *testing.B) {
	var payload interface{}
	payload = "hello world"
	benchmarkAddEvents(100, payload, b)
}

func BenchmarkAddEvents1000string(b *testing.B) {
	var payload interface{}
	payload = "hello world"
	benchmarkAddEvents(1000, payload, b)
}

func BenchmarkAddEvents100arr(b *testing.B) {
	var payload interface{}
	payload = []string{"hello", "world"}
	benchmarkAddEvents(100, payload, b)
}

func BenchmarkAddEvents1000arr(b *testing.B) {
	var payload interface{}
	payload = []string{"hello", "world"}
	benchmarkAddEvents(1000, payload, b)
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
