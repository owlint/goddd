package goddd

import "testing"

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

func (object *testDomainObject) Apply(eventName string, event interface{}) {
	switch eventName {
	case "NumberAdded":
		toAdd := event.(int32)
		object.number += toAdd
	}
}

func (object *testDomainObject) Method(number int32) {
	object.AddEvent(object, "NumberAdded", number)
}

func TestMutationWorks(t *testing.T) {
	object := testDomainObject{}

	object.Method(5)

	if object.number != 5 {
		t.Error("Wrong value")
		t.FailNow()
	}

	if len(object.Events()) != 1 {
		t.Error("Cant find events")
		t.FailNow()
	}
}

func TestMultipleMutationWorks(t *testing.T) {
	object := testDomainObject{}

	object.Method(5)
	object.Method(5)

	if object.number != 10 {
		t.Error("Wrong value")
		t.FailNow()
	}

	if len(object.Events()) != 2 {
		t.Error("Cant find events")
		t.FailNow()
	}
}
