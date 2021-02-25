package goddd

import "testing"

func TestSave(t *testing.T) {
	publisher := NewEventPublisher()
	repo := NewInMemoryRepository(publisher)
	object := testDomainObject{}

	object.Method(5)
	repo.Save(&object)

	if len(repo.eventStream) != 1 {
		t.Error("Should contain only one event")
		t.FailNow()
	}
}

func TestSaveMultiples(t *testing.T) {
	publisher := NewEventPublisher()
	repo := NewInMemoryRepository(publisher)
	object := testDomainObject{}

	object.Method(5)
	object.Method(7)
	repo.Save(&object)

	if len(repo.eventStream) != 2 {
		t.Error("Should contain only one event")
		t.FailNow()
	}
}

func TestSaveMultiplesObject(t *testing.T) {
	publisher := NewEventPublisher()
	repo := NewInMemoryRepository(publisher)
	object := testDomainObject{ID: NewIdentity("TestDomainObject")}
	object2 := testDomainObject{ID: NewIdentity("TestDomainObject")}

	object.Method(5)
	object.Method(7)
	repo.Save(&object)

	object2.Method(10)
	repo.Save(&object2)

	if len(repo.eventStream) != 3 {
		t.Errorf("Should contain only one event : %d", len(repo.eventStream))
		t.FailNow()
	}
}

func TestLoad(t *testing.T) {
	publisher := NewEventPublisher()
	repo := NewInMemoryRepository(publisher)
	object := testDomainObject{ID: NewIdentity("TestDomainObject")}
	object2 := testDomainObject{ID: NewIdentity("TestDomainObject")}

	object.Method(5)
	object.Method(7)
	repo.Save(&object)

	object2.Method(10)
	repo.Save(&object2)

	loadedObject := testDomainObject{}
	err := repo.Load(object.ObjectID(), &loadedObject)

	if err != nil {
		t.Errorf("No error should occur on load : %v", err)
		t.FailNow()
	}

	if loadedObject.number != 12 {
		t.Error("Wrong value")
		t.FailNow()
	}
}
