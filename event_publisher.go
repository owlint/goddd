package goddd

import "sync"

type EventReceiver interface {
	OnEvent(event Event)
}

type EventPublisher struct {
	receivers []EventReceiver
	Wait      bool
}

func (p *EventPublisher) Register(receiver EventReceiver) {
	p.receivers = append(p.receivers, receiver)
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
