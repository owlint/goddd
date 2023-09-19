package goddd

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSave(t *testing.T) {
	publisher := NewEventPublisher()
	repo := NewInMemoryRepository[*Student](&publisher)
	object := Student{}

	object.SetGrade("a")
	repo.Save(context.Background(), &object)

	if len(repo.eventStream) != 1 {
		t.Error("Should contain only one event")
		t.FailNow()
	}
}

func TestSaveMultiples(t *testing.T) {
	publisher := NewEventPublisher()
	repo := NewInMemoryRepository[*Student](&publisher)
	object := Student{}

	object.SetGrade("a")
	object.SetGrade("a")
	repo.Save(context.Background(), &object)

	if len(repo.eventStream) != 2 {
		t.Error("Should contain only one event")
		t.FailNow()
	}
}

func TestSaveMultiplesObject(t *testing.T) {
	publisher := NewEventPublisher()
	repo := NewInMemoryRepository[*Student](&publisher)
	object := Student{}
	object2 := Student{}

	object.SetGrade("a")
	object.SetGrade("a")
	repo.Save(context.Background(), &object)

	object2.SetGrade("a")
	repo.Save(context.Background(), &object2)

	if len(repo.eventStream) != 3 {
		t.Errorf("Should contain only one event : %d", len(repo.eventStream))
		t.FailNow()
	}
}

func TestLoad(t *testing.T) {
	publisher := NewEventPublisher()
	repo := NewInMemoryRepository[*Student](&publisher)
	object := Student{}
	object2 := Student{}

	object.SetGrade("a")
	object.SetGrade("a")
	repo.Save(context.Background(), &object)

	object2.SetGrade("a")
	repo.Save(context.Background(), &object2)

	loadedObject := Student{}
	err := repo.Load(context.Background(), object.ObjectID(), &loadedObject)

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
	repo := NewInMemoryRepository[*Student](&publisher)
	object := Student{}
	object2 := Student{}

	object.SetGrade("a")
	repo.Save(context.Background(), &object)
	object2.SetGrade("a")
	repo.Save(context.Background(), &object2)
	before := time.Now()

	time.Sleep(3 * time.Second)

	object.SetGrade("b")
	time.Sleep(100 * time.Millisecond)
	object2.SetGrade("b")
	repo.Save(context.Background(), &object)
	repo.Save(context.Background(), &object2)

	events, err := repo.EventsSince(context.Background(), before, 50)

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
	repo := NewInMemoryRepository[*Student](&publisher)
	object := Student{}
	object2 := Student{}

	object.SetGrade("a")
	repo.Save(context.Background(), &object)
	object2.SetGrade("a")
	repo.Save(context.Background(), &object2)
	before := time.Now()

	time.Sleep(3 * time.Second)

	object.SetGrade("b")
	time.Sleep(100 * time.Millisecond)
	object2.SetGrade("b")
	repo.Save(context.Background(), &object)
	repo.Save(context.Background(), &object2)

	events, err := repo.EventsSince(context.Background(), before, 1)

	if err != nil {
		t.Error("No error should occur on EventsSince")
		t.FailNow()
	}

	if len(events) != 1 {
		t.Errorf("There should have 1 event, got %d : %v", len(events), events)
		t.FailNow()
	}
}
func TestInMemoryRemove(t *testing.T) {
	t.Run("Remove", func(t *testing.T) {
		publisher := NewEventPublisher()
		repo := NewInMemoryRepository[*Student](&publisher)
		object := Student{ID: uuid.New().String()}
		object.SetGrade("a")
		err := repo.Save(context.Background(), &object)
		assert.NoError(t, err)

		err = repo.Remove(context.Background(), object.ObjectID(), &object)

		assert.NoError(t, err)
		exists, err := repo.Exists(context.Background(), object.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		events := make([]Event, 0)
		for _, event := range repo.eventStream {
			if event.objectID == object.ID {
				events = append(events, event)
			}
		}

		assert.Len(t, events, 1)
		assert.Equal(t, REMOVED_EVENT_NAME, events[0].name)
	})
	t.Run("Remove unknown", func(t *testing.T) {
		publisher := NewEventPublisher()
		repo := NewInMemoryRepository[*Student](&publisher)
		object := Student{ID: uuid.New().String()}
		object.SetGrade("a")
		err := repo.Save(context.Background(), &object)
		assert.NoError(t, err)

		err = repo.Remove(context.Background(), uuid.NewString(), &object)

		assert.Error(t, err)
		exists, err := repo.Exists(context.Background(), object.ID)
		assert.NoError(t, err)
		assert.True(t, exists)
	})
}
