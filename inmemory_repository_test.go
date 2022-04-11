package goddd

import (
	"testing"
	"time"
)

func TestSave(t *testing.T) {
	publisher := NewEventPublisher()
	repo := NewInMemoryRepository(&publisher)
	object := Student{}

	object.SetGrade("a")
	repo.Save(&object)

	if len(repo.eventStream) != 1 {
		t.Error("Should contain only one event")
		t.FailNow()
	}
}

func TestSaveMultiples(t *testing.T) {
	publisher := NewEventPublisher()
	repo := NewInMemoryRepository(&publisher)
	object := Student{}

	object.SetGrade("a")
	object.SetGrade("a")
	repo.Save(&object)

	if len(repo.eventStream) != 2 {
		t.Error("Should contain only one event")
		t.FailNow()
	}
}

func TestSaveMultiplesObject(t *testing.T) {
	publisher := NewEventPublisher()
	repo := NewInMemoryRepository(&publisher)
	object := Student{}
	object2 := Student{}

	object.SetGrade("a")
	object.SetGrade("a")
	repo.Save(&object)

	object2.SetGrade("a")
	repo.Save(&object2)

	if len(repo.eventStream) != 3 {
		t.Errorf("Should contain only one event : %d", len(repo.eventStream))
		t.FailNow()
	}
}

func TestLoad(t *testing.T) {
	publisher := NewEventPublisher()
	repo := NewInMemoryRepository(&publisher)
	object := Student{}
	object2 := Student{}

	object.SetGrade("a")
	object.SetGrade("a")
	repo.Save(&object)

	object2.SetGrade("a")
	repo.Save(&object2)

	loadedObject := Student{}
	err := repo.Load(object.ObjectID(), &loadedObject)

	if err != nil {
		t.Errorf("No error should occur on load : %v", err)
		t.FailNow()
	}

	if loadedObject.grade != "a" {
		t.Error("Wrong value")
		t.FailNow()
	}
}

func TestInMemoryEventsSince(t *testing.T) {
	publisher := NewEventPublisher()
	repo := NewInMemoryRepository(&publisher)
	object := Student{}
	object2 := Student{}

	object.SetGrade("a")
	repo.Save(&object)
	object2.SetGrade("a")
	repo.Save(&object2)
	before := time.Now()

	time.Sleep(3 * time.Second)

	object.SetGrade("b")
	time.Sleep(100 * time.Millisecond)
	object2.SetGrade("b")
	repo.Save(&object)
	repo.Save(&object2)

	events, err := repo.EventsSince(before, 50)

	if err != nil {
		t.Error("No error should occur on EventsSince")
		t.FailNow()
	}

	if len(events) != 2 {
		t.Errorf("There should have 2 events, got %d : %v", len(events), events)
		t.FailNow()
	}
}

func TestInMemoryEventsSinceLimit(t *testing.T) {

	publisher := NewEventPublisher()
	repo := NewInMemoryRepository(&publisher)
	object := Student{}
	object2 := Student{}

	object.SetGrade("a")
	repo.Save(&object)
	object2.SetGrade("a")
	repo.Save(&object2)
	before := time.Now()

	time.Sleep(3 * time.Second)

	object.SetGrade("b")
	time.Sleep(100 * time.Millisecond)
	object2.SetGrade("b")
	repo.Save(&object)
	repo.Save(&object2)

	events, err := repo.EventsSince(before, 1)

	if err != nil {
		t.Error("No error should occur on EventsSince")
		t.FailNow()
	}

	if len(events) != 1 {
		t.Errorf("There should have 1 event, got %d : %v", len(events), events)
		t.FailNow()
	}
}
