package goddd

import (
	"testing"

	"github.com/owlint/goddd/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

type testDomainObject struct {
	Stream

	ID     string
	number int32
}

func (object *testDomainObject) ObjectID() string {
	if object.ID == "" {
		return "objectId"
	}
	return object.ID
}

func (object *testDomainObject) Apply(eventName string, event []byte) error {
	switch eventName {
	case "NumberAdded":
		payload := &pb.NumberAdded{}
		err := proto.Unmarshal(event, payload)
		if err != nil {
			return err
		}
		object.number += payload.Nb
		return nil
	}

	return nil
}

func (object *testDomainObject) Method(number int32) {
	event := pb.NumberAdded{}
	event.Nb = number
	object.AddEvent(object, "NumberAdded", &event)
}

func TestMutationWorks(t *testing.T) {
	object := testDomainObject{}

	object.Method(5)

	assert.Equal(t, int32(5), object.number)
	assert.Len(t, object.Events(), 1)
}

func TestMultipleMutationWorks(t *testing.T) {
	object := testDomainObject{}

	object.Method(5)
	object.Method(5)

	assert.Equal(t, int32(10), object.number)
	assert.Len(t, object.Events(), 2)
}
