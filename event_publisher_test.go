package goddd

import (
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/golang/mock/gomock"
	"github.com/owlint/goddd/mocks"
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
	event := NewEvent("TestObject", "name", 1, []byte{1, 2})
	publisher.Publish([]Event{event})

	assert.Len(t, receiver.events, 1)
	assertEventsEqual(t, event, receiver.events[0])

}

func TestOnEvent(t *testing.T) {
	publisher := NewEventPublisher()
	publisher.Wait = true
	receiver := testReceiver{
		events: make([]Event, 0),
	}

	publisher.Register(&receiver)
	event := NewEvent("TestObject", "name", 1, []byte{1, 2})
	publisher.OnEvent(event)

	assert.Len(t, receiver.events, 1)
	assertEventsEqual(t, event, receiver.events[0])
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

	event := NewEvent("TestObject", "name", 1, []byte{1, 2})
	publisher.Publish([]Event{event})

	assert.Len(t, receiver1.events, 1)
	assertEventsEqual(t, event, receiver1.events[0])
	assert.Len(t, receiver2.events, 1)
	assertEventsEqual(t, event, receiver2.events[0])
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
		event := NewEvent("TestObject", "name", 1, []byte{1, 2})
		remotePublisher.OnEvent(event)

		// HACK: wait for event to be processed by listener
		time.Sleep(time.Second)

		assert.Len(t, errChan, 0)
		assert.Len(t, receiver.events, 1)
		assertEventsEqual(t, event, receiver.events[0])
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
		event1 := NewEvent("TestObject", "name", 1, []byte{1, 2})
		event2 := NewEvent("TestObject2", "name", 1, []byte{1, 2})
		remotePublisher.OnEvent(event1)
		remotePublisher.OnEvent(event2)

		// HACK: wait for events to be processed by listener
		time.Sleep(time.Second)

		assert.Len(t, errChan, 0)
		assert.Len(t, receiver.events, 2)
		assertEventsEqual(t, event1, receiver.events[0])
		assertEventsEqual(t, event2, receiver.events[1])
	})
}

func TestRemotePublishQueueError(t *testing.T) {
	ctrl := gomock.NewController(t)
	queue := mocks.NewMockQueueService(ctrl)

	err := errors.New("bam")
	queue.EXPECT().Push(gomock.Any(), gomock.Any()).Return(err)

	errChan := make(chan error, 1)
	remotePublisher := NewRemoteEventPublisher(queue, errChan)

	remotePublisher.OnEvent(NewEvent("TestObject", "name", 1, []byte{1, 2}))
	assert.Len(t, errChan, 1)
	assert.Equal(t, err, <-errChan)
}

func TestRemoteListenerQueueError(t *testing.T) {
	ctrl := gomock.NewController(t)
	queue := mocks.NewMockQueueService(ctrl)

	err := errors.New("boum")
	queue.EXPECT().Pop(gomock.Any()).AnyTimes().Return(nil, err)

	errChan := make(chan error, 1)
	receiver := testReceiver{
		events: make([]Event, 0),
	}

	remoteListener := NewRemoteEventListener(queue, &receiver, errChan)

	go remoteListener.Listen()

	// // HACK: wait for listener
	time.Sleep(time.Second)

	assert.Len(t, errChan, 1)
	assert.Equal(t, err, <-errChan)
}
