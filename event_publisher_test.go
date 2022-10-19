package goddd

import (
	"testing"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/owlint/goddd/services"
	"github.com/owlint/goddd/testutils"
	"github.com/stretchr/testify/assert"
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

	assert.Len(t, receiver.events, 1)
}

func TestOnEvent(t *testing.T) {
	publisher := NewEventPublisher()
	publisher.Wait = true
	receiver := testReceiver{
		events: make([]Event, 0),
	}

	publisher.Register(&receiver)

	publisher.OnEvent(NewEvent("TestObject", "name", 1, []byte{1, 2}))

	assert.Len(t, receiver.events, 1)
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

	assert.Len(t, receiver1.events, 1)
	assert.Len(t, receiver2.events, 1)
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

	assert.Len(t, receiver1.events, 200)
	assert.Len(t, receiver2.events, 200)
}

func TestRemotePublish(t *testing.T) {
	testutils.WithTestRedis(func(conn *redis.Client) {
		queue := services.NewRedisQueueService(conn, "test")

		errChan := make(chan error, 1)
		receiver := testReceiver{
			events: make([]Event, 0),
		}

		remotePublisher := NewRemoteEventPublisher(queue, errChan)
		remoteListener := NewRemoteEventListener(queue, &receiver, errChan)

		go remoteListener.Listen()
		remotePublisher.OnEvent(NewEvent("TestObject", "name", 1, []byte{1, 2}))

		// HACK: wait for event to be processed by listener
		time.Sleep(time.Second)

		assert.Len(t, errChan, 0)
		assert.Len(t, receiver.events, 1)
	})
}

func TestRemotePublishMultiple(t *testing.T) {
	testutils.WithTestRedis(func(conn *redis.Client) {
		queue := services.NewRedisQueueService(conn, "test")

		errChan := make(chan error, 1)
		receiver := testReceiver{
			events: make([]Event, 0),
		}

		remotePublisher := NewRemoteEventPublisher(queue, errChan)
		remoteListener := NewRemoteEventListener(queue, &receiver, errChan)

		go remoteListener.Listen()
		remotePublisher.OnEvent(NewEvent("TestObject", "name", 1, []byte{1, 2}))
		remotePublisher.OnEvent(NewEvent("TestObject2", "name", 1, []byte{1, 2}))

		// HACK: wait for events to be processed by listener
		time.Sleep(time.Second)

		assert.Len(t, errChan, 0)
		assert.Len(t, receiver.events, 2)
	})
}
