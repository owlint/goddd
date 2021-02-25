package goddd

import (
	"testing"
)

type testReceiver struct {
	events []Event
}

func (r *testReceiver) OnEvent(event Event) {
	r.events = append(r.events, event)
}

func TestPublished(t *testing.T) {
	publisher := NewEventPublisher()
	publisher.Wait = true
	receiver := testReceiver{
		events: make([]Event, 0),
	}

	publisher.Register(&receiver)

	publisher.Publish([]Event{NewEvent("TestObject", "name", 1, []byte{1, 2})})

	if len(receiver.events) != 1 {
		t.Error("Wrong number of events")
	}

}

func TestMultipleReceivers(t *testing.T) {
	publisher := NewEventPublisher()
	publisher.Wait = true
	receiver1 := testReceiver{
		events: make([]Event, 0),
	}
	receiver2 := testReceiver{
		events: make([]Event, 0),
	}

	publisher.Register(&receiver1)
	publisher.Register(&receiver2)

	publisher.Publish([]Event{NewEvent("TestObject", "name", 1, []byte{1, 2})})

	if len(receiver1.events) != 1 {
		t.Errorf("Wrong number of events, %d", len(receiver1.events))
	}

	if len(receiver2.events) != 1 {
		t.Errorf("Wrong number of events, %d", len(receiver2.events))
	}
}

func TestMultiplePublishAndReceivers(t *testing.T) {
	publisher := NewEventPublisher()
	publisher.Wait = true
	receiver1 := testReceiver{
		events: make([]Event, 0),
	}
	receiver2 := testReceiver{
		events: make([]Event, 0),
	}

	publisher.Register(&receiver1)
	publisher.Register(&receiver2)

	for i := 0; i < 100; i++ {
		publisher.Publish([]Event{NewEvent("TestObject", "name", 1, []byte{3, 4}), NewEvent("TestObject", "another", 2, []byte{1, 2})})
	}

	if len(receiver1.events) != 200 {
		t.Errorf("Wrong number of events, %d", len(receiver1.events))
	}

	if len(receiver2.events) != 200 {
		t.Errorf("Wrong number of events, %d", len(receiver2.events))
	}
}
