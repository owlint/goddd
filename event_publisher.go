package goddd

import (
	"context"
	"sync"

	"github.com/owlint/goddd/services"
)

type EventReceiver interface {
	OnEvent(event Event)
}

type EventPublisher struct {
	receivers []EventReceiver
	Wait      bool
}

type RemoteEventPublisher struct {
	queue   services.QueueService
	errChan chan<- error
}

type RemoteEventListener struct {
	queue    services.QueueService
	receiver EventReceiver
	errChan  chan<- error
}

func (p *EventPublisher) Register(receiver EventReceiver) {
	p.receivers = append(p.receivers, receiver)
}

func (p *EventPublisher) OnEvent(event Event) {
	p.Publish([]Event{event})
}

func (p *EventPublisher) Publish(events []Event) {
	var wg sync.WaitGroup

	for _, receiver := range p.receivers {

		if p.Wait {
			wg.Add(1)
		}

		go func(receiver EventReceiver) {
			for _, event := range events {
				receiver.OnEvent(event)
			}

			if p.Wait {
				wg.Done()
			}
		}(receiver)
	}

	if p.Wait {
		wg.Wait()
	}
}

func NewEventPublisher() EventPublisher {
	return EventPublisher{
		receivers: make([]EventReceiver, 0),
		Wait:      false,
	}
}

func NewRemoteEventPublisher(queue services.QueueService, errChan chan<- error) EventReceiver {
	return &RemoteEventPublisher{
		queue:   queue,
		errChan: errChan,
	}
}

func (r *RemoteEventPublisher) OnEvent(event Event) {
	serializedEvent, err := event.Serialize()
	if err != nil {
		r.errChan <- err
		return
	}

	err = r.queue.Push(context.TODO(), serializedEvent)
	if err != nil {
		r.errChan <- err
		return
	}
}

func NewRemoteEventListener(queue services.QueueService, receiver EventReceiver, errChan chan<- error) RemoteEventListener {
	return RemoteEventListener{
		queue:    queue,
		receiver: receiver,
		errChan:  errChan,
	}
}

func (r *RemoteEventListener) Listen() {
	for {
		serialized, err := r.queue.Pop(context.TODO())
		if err != nil {
			r.errChan <- err
			continue
		}

		event, err := Deserialize(serialized)
		if err != nil {
			r.errChan <- err
			continue
		}

		r.receiver.OnEvent(event)
	}
}
