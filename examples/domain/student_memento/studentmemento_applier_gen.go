package student_memento

import (
	"errors"
)

// This code have been generated. DO NOT MODIFY

func (z *StudentMemento) Apply(eventName string, eventPayload []byte) error {
	switch eventName {
	case "GradeSet":
		event := GradeSet{}
		_, err := event.UnmarshalMsg(eventPayload)
		if err != nil {
			return err
		}
		return z.OnGradeSet(event)

	default:
		return errors.New("Unknown event type")
	}
}

func (z *StudentMemento) ApplyMemento(payload []byte) error {
	memento := Memento{}
	_, err := memento.UnmarshalMsg(payload)
	if err != nil {
		return err
	}
	return z.ReloadMemento(memento)
}

