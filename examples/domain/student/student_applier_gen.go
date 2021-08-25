
package student

import (
    "errors"
)

// This code have been generated. DO NOT MODIFY

func (z *Student) Apply(eventName string, eventPayload []byte) error {
    switch(eventName) {
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
        