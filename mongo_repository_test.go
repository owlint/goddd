package goddd

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func connectTestMongo(t *testing.T) (*mongo.Client, *mongo.Database) {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, e := mongo.Connect(context.TODO(), clientOptions)
	if e != nil {
		t.Error(e)
		t.Fail()
	}

	// Check the connection
	e = client.Ping(context.TODO(), nil)
	if e != nil {
		t.Error(e)
		t.Fail()
	}

	// get collection as ref
	database := client.Database("testdb")
	return client, database
}

func eventStreamFor(t *testing.T, database *mongo.Database, objectID string) []Event {
	collection := database.Collection("event_store")

	stream := make([]record, 0)
	cursor, err := collection.Find(context.TODO(), bson.D{{"objectid", objectID}})

	if err != nil {
		t.Error(err)
	}

	err = cursor.All(context.TODO(), &stream)
	if err != nil {
		t.Error(err)
	}

	return fromRecords(stream)
}

func TestMongoSave(t *testing.T) {
	t.Run("With valid data", func(t *testing.T) {
		client, database := connectTestMongo(t)
		defer client.Disconnect(context.TODO())

		publisher := NewEventPublisher()
		repo, err := NewMongoRepository[*Student](database, &publisher)
		assert.NoError(t, err)
		object := Student{ID: uuid.NewString()}

		object.SetGrade("a")
		err = repo.Save(context.Background(), &object)
		assert.NoError(t, err)

		assert.Len(t, eventStreamFor(t, database, object.ObjectID()), 1)
	})
	t.Run("save is idempotent", func(t *testing.T) {
		client, database := connectTestMongo(t)
		defer client.Disconnect(context.TODO())

		publisher := NewEventPublisher()
		repo, err := NewMongoRepository[*Student](database, &publisher)
		assert.NoError(t, err)
		object := Student{ID: uuid.NewString()}

		object.SetGrade("a")
		err = repo.Save(context.Background(), &object)
		assert.NoError(t, err)
		assert.Len(t, eventStreamFor(t, database, object.ObjectID()), 1)

		object.SetGrade("a")
		err = repo.Save(context.Background(), &object)
		assert.NoError(t, err)
		assert.Len(t, eventStreamFor(t, database, object.ObjectID()), 2)
		err = repo.Save(context.Background(), &object)
		assert.NoError(t, err)
		assert.Len(t, eventStreamFor(t, database, object.ObjectID()), 2)
	})

	t.Run("Concurrency error", func(t *testing.T) {
		client, database := connectTestMongo(t)
		defer client.Disconnect(context.TODO())

		publisher := NewEventPublisher()
		repo, err := NewMongoRepository[*Student](database, &publisher)
		assert.NoError(t, err)
		object := Student{ID: uuid.NewString()}

		object.SetGrade("a")
		err = repo.Save(context.Background(), &object)
		assert.NoError(t, err)

		objectCopy := object
		object.SetGrade("b")
		err = repo.Save(context.Background(), &object)
		assert.NoError(t, err)
		objectCopy.SetGrade("c")
		err = repo.Save(context.Background(), &objectCopy)
		assert.ErrorIs(t, err, ConcurrencyError)
	})
}

func TestMongoPublished(t *testing.T) {
	client, database := connectTestMongo(t)
	defer client.Disconnect(context.TODO())

	receiver := testReceiver{
		events: make([]Event, 0),
	}
	publisher := NewEventPublisher()
	publisher.Register(&receiver)
	publisher.Wait = true
	repo, err := NewMongoRepository[*Student](database, &publisher)
	assert.NoError(t, err)
	object := Student{}

	object.SetGrade("a")
	err = repo.Save(context.Background(), &object)
	assert.NoError(t, err)

	if len(receiver.events) != 1 {
		t.Error("No events received")
		t.FailNow()
	}
}

func TestMongoPublishedNoWait(t *testing.T) {
	client, database := connectTestMongo(t)
	defer client.Disconnect(context.TODO())

	receiver := testReceiver{
		events: make([]Event, 0),
	}
	publisher := NewEventPublisher()
	publisher.Register(&receiver)
	repo, err := NewMongoRepository[*Student](database, &publisher)
	assert.NoError(t, err)
	object := Student{
		ID: uuid.New().String(),
	}

	object.SetGrade("a")
	err = repo.Save(context.Background(), &object)
	assert.NoError(t, err)
	time.Sleep(500 * time.Millisecond)

	if len(receiver.events) != 1 {
		t.Error("No events received")
		t.FailNow()
	}
}

func TestMongoSaveMultiples(t *testing.T) {
	client, database := connectTestMongo(t)
	defer client.Disconnect(context.TODO())

	publisher := NewEventPublisher()
	repo, err := NewMongoRepository[*Student](database, &publisher)
	assert.NoError(t, err)
	object := Student{ID: uuid.New().String()}

	object.SetGrade("a")
	object.SetGrade("a")
	err = repo.Save(context.Background(), &object)
	assert.NoError(t, err)

	if len(eventStreamFor(t, database, object.ObjectID())) != 2 {
		t.Error("Should contain only one event")
		t.FailNow()
	}
}

func TestMongoSaveMultiplesObject(t *testing.T) {
	client, database := connectTestMongo(t)
	defer client.Disconnect(context.TODO())

	publisher := NewEventPublisher()
	repo, err := NewMongoRepository[*Student](database, &publisher)
	assert.NoError(t, err)

	object := Student{ID: uuid.New().String()}
	object2 := Student{ID: uuid.New().String()}

	object.SetGrade("a")
	object.SetGrade("a")
	err = repo.Save(context.Background(), &object)
	assert.NoError(t, err)

	object2.SetGrade("a")
	err = repo.Save(context.Background(), &object2)
	assert.NoError(t, err)

	if len(eventStreamFor(t, database, object.ObjectID())) != 2 {
		t.Errorf("Should contain only one event : %d", len(eventStreamFor(t, database, object.ObjectID())))
		t.FailNow()
	}
}

func TestMongoLoad(t *testing.T) {
	client, database := connectTestMongo(t)
	defer client.Disconnect(context.TODO())

	publisher := NewEventPublisher()
	repo, err := NewMongoRepository[*Student](database, &publisher)
	assert.NoError(t, err)
	object := Student{ID: uuid.New().String()}
	object2 := Student{ID: uuid.New().String()}

	object.SetGrade("a")
	object.SetGrade("a")
	err = repo.Save(context.Background(), &object)
	assert.NoError(t, err)

	object2.SetGrade("a")
	err = repo.Save(context.Background(), &object2)
	assert.NoError(t, err)

	loadedObject := Student{}
	err = repo.Load(context.Background(), object.ObjectID(), &loadedObject)

	if err != nil {
		t.Errorf("No error should occur on load : %v", err)
		t.FailNow()
	}

	if loadedObject.grade != "a" {
		t.Error("Wrong value")
		t.FailNow()
	}
}

func TestMongoEventsSince(t *testing.T) {
	client, database := connectTestMongo(t)
	defer client.Disconnect(context.TODO())

	publisher := NewEventPublisher()
	repo, err := NewMongoRepository[*Student](database, &publisher)
	assert.NoError(t, err)
	object := Student{ID: uuid.New().String()}
	object2 := Student{ID: uuid.New().String()}

	object.SetGrade("a")
	err = repo.Save(context.Background(), &object)
	assert.NoError(t, err)
	object2.SetGrade("a")
	err = repo.Save(context.Background(), &object2)
	assert.NoError(t, err)
	before := time.Now()

	time.Sleep(3 * time.Second)

	object.SetGrade("b")
	time.Sleep(100 * time.Millisecond)
	object2.SetGrade("b")
	err = repo.Save(context.Background(), &object)
	assert.NoError(t, err)
	err = repo.Save(context.Background(), &object2)
	assert.NoError(t, err)

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

func TestMongoEventsSinceLimit(t *testing.T) {
	client, database := connectTestMongo(t)
	defer client.Disconnect(context.TODO())

	publisher := NewEventPublisher()
	repo, err := NewMongoRepository[*Student](database, &publisher)
	assert.NoError(t, err)
	object := Student{ID: uuid.New().String()}
	object2 := Student{ID: uuid.New().String()}

	object.SetGrade("a")
	err = repo.Save(context.Background(), &object)
	assert.NoError(t, err)
	object2.SetGrade("a")
	err = repo.Save(context.Background(), &object2)
	assert.NoError(t, err)
	before := time.Now()

	time.Sleep(3 * time.Second)

	object.SetGrade("b")
	time.Sleep(100 * time.Millisecond)
	object2.SetGrade("b")
	err = repo.Save(context.Background(), &object)
	assert.NoError(t, err)
	err = repo.Save(context.Background(), &object2)
	assert.NoError(t, err)

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

func TestMongoMementizerSnapshotSaved(t *testing.T) {
	client, database := connectTestMongo(t)
	defer client.Disconnect(context.TODO())

	publisher := NewEventPublisher()
	repo, err := NewMongoRepository[*StudentMemento](database, &publisher)

	assert.NoError(t, err)
	object := StudentMemento{
		EventStream: &Stream{},
		ID:          uuid.New().String(),
	}

	for i := 0; i < 600; i++ {
		object.SetGrade(fmt.Sprintf("a%d", i))
	}
	err = repo.Save(context.Background(), &object)
	assert.NoError(t, err)

	filter := bson.M{"objectid": object.ObjectID()}
	result := database.Collection("domain_event_snapshots").FindOne(context.Background(), filter)

	if result.Err() != nil {
		t.Errorf("There should have a snapshot for this object but got an error : %s", result.Err().Error())
		t.FailNow()
	}
}

func TestMongoNotMementizerSnapshotNotSaved(t *testing.T) {
	client, database := connectTestMongo(t)
	defer client.Disconnect(context.TODO())

	publisher := NewEventPublisher()
	repo, err := NewMongoRepository[*Student](database, &publisher)
	assert.NoError(t, err)
	object := Student{ID: uuid.New().String()}

	for i := 0; i < 600; i++ {
		object.SetGrade(fmt.Sprintf("a%d", i))
	}
	err = repo.Save(context.Background(), &object)
	assert.NoError(t, err)

	filter := bson.M{"objectid": object.ObjectID()}
	result := database.Collection("domain_event_snapshots").FindOne(context.Background(), filter)

	if result.Err() == nil || result.Err() != mongo.ErrNoDocuments {
		t.Error("There should have a snapshot for this domain object")
		panic(result.Err())
	}
}

func TestMongoLoadMementizer(t *testing.T) {
	client, database := connectTestMongo(t)
	defer client.Disconnect(context.TODO())

	publisher := NewEventPublisher()
	repo, err := NewMongoRepository[*StudentMemento](database, &publisher)
	assert.NoError(t, err)
	object := StudentMemento{
		EventStream: &Stream{},
		ID:          uuid.New().String(),
	}

	for i := 0; i < 600; i++ {
		object.SetGrade(fmt.Sprintf("a%d", i))
	}
	err = repo.Save(context.Background(), &object)
	assert.NoError(t, err)
	for i := 0; i < 10; i++ {
		object.SetGrade(fmt.Sprintf("a%d", i))
	}
	err = repo.Save(context.Background(), &object)
	assert.NoError(t, err)

	loadedObject := StudentMemento{
		EventStream: &Stream{},
	}
	err = repo.Load(context.Background(), object.ObjectID(), &loadedObject)

	if err != nil {
		t.Error("Loading should not throw an error")
		t.FailNow()
	}

	if object.ObjectID() != loadedObject.ObjectID() {
		t.Error("Different Object IDs")
		t.FailNow()
	}

	if object.grade != loadedObject.grade {
		t.Error("Different grades")
		t.FailNow()
	}

	if object.LastVersion() != loadedObject.LastVersion() {
		t.Errorf("Different LastVersion, %d, %d", object.LastVersion(), loadedObject.LastVersion())
		t.FailNow()
	}
}

func TestMongoUpdate(t *testing.T) {
	t.Run("Update", func(t *testing.T) {
		client, database := connectTestMongo(t)
		defer client.Disconnect(context.TODO())

		publisher := NewEventPublisher()
		repo, err := NewMongoRepository[*Student](database, &publisher)
		assert.NoError(t, err)
		object := Student{ID: uuid.New().String()}
		object.SetGrade("a")
		err = repo.Save(context.Background(), &object)
		assert.NoError(t, err)

		savedObject, err := repo.Update(context.Background(), object.ObjectID(), &object, 1, func(object *Student) (*Student, error) {
			object.SetGrade("b")
			return object, nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "b", savedObject.grade)

		loadedObject := &Student{}
		err = repo.Load(context.Background(), object.ObjectID(), loadedObject)
		assert.NoError(t, err)
		assert.Equal(t, "b", loadedObject.grade)
	})

	t.Run("Retry on ConcurrencyError", func(t *testing.T) {
		client, database := connectTestMongo(t)
		defer client.Disconnect(context.TODO())

		publisher := NewEventPublisher()
		repo, err := NewMongoRepository[*Student](database, &publisher)
		assert.NoError(t, err)
		object := Student{ID: uuid.New().String()}
		object.SetGrade("a")
		err = repo.Save(context.Background(), &object)
		assert.NoError(t, err)

		// Here we introduce a concurrency error
		object2 := object
		object2.SetGrade("c")
		err = repo.Save(context.Background(), &object2)
		assert.NoError(t, err)

		// Now we update and should not get an error
		savedObject, err := repo.Update(context.Background(), object.ObjectID(), &object, 1, func(object *Student) (*Student, error) {
			object.SetGrade("b")
			return object, nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "b", savedObject.grade)

		loadedObject := &Student{}
		err = repo.Load(context.Background(), object.ObjectID(), loadedObject)
		assert.NoError(t, err)
		assert.Equal(t, "b", loadedObject.grade)
	})

	t.Run("Invalid retries", func(t *testing.T) {
		client, database := connectTestMongo(t)
		defer client.Disconnect(context.TODO())

		publisher := NewEventPublisher()
		repo, err := NewMongoRepository[*Student](database, &publisher)
		assert.NoError(t, err)
		object := Student{ID: uuid.New().String()}
		object.SetGrade("a")
		err = repo.Save(context.Background(), &object)
		assert.NoError(t, err)

		_, err = repo.Update(context.Background(), object.ObjectID(), &object, -1, func(object *Student) (*Student, error) {
			object.SetGrade("b")
			return object, nil
		})

		assert.Error(t, err)
	})

	t.Run("Update error", func(t *testing.T) {
		client, database := connectTestMongo(t)
		defer client.Disconnect(context.TODO())

		publisher := NewEventPublisher()
		repo, err := NewMongoRepository[*Student](database, &publisher)
		assert.NoError(t, err)
		object := Student{ID: uuid.New().String()}
		object.SetGrade("a")
		err = repo.Save(context.Background(), &object)
		assert.NoError(t, err)

		myError := errors.New("my error")
		_, err = repo.Update(context.Background(), object.ObjectID(), &object, 0, func(object *Student) (*Student, error) {
			return object, myError
		})

		assert.Error(t, err)
		assert.True(t, errors.Is(err, myError))
	})

	t.Run("No error nil object", func(t *testing.T) {
		client, database := connectTestMongo(t)
		defer client.Disconnect(context.TODO())

		publisher := NewEventPublisher()
		repo, err := NewMongoRepository[*Student](database, &publisher)
		assert.NoError(t, err)
		object := Student{ID: uuid.New().String()}
		object.SetGrade("a")
		err = repo.Save(context.Background(), &object)
		assert.NoError(t, err)

		_, err = repo.Update(context.Background(), object.ObjectID(), &object, 1, func(object *Student) (*Student, error) {
			object.SetGrade("b")
			return nil, nil
		})

		assert.Error(t, err)
		assert.ErrorIs(t, err, InvalidUpdateCallback)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		client, database := connectTestMongo(t)
		defer client.Disconnect(context.TODO())

		publisher := NewEventPublisher()
		repo, err := NewMongoRepository[*Student](database, &publisher)
		assert.NoError(t, err)
		object := Student{ID: uuid.New().String()}
		object.SetGrade("a")
		err = repo.Save(context.Background(), &object)
		assert.NoError(t, err)

		_, err = repo.Update(context.Background(), uuid.NewString(), &object, -1, func(object *Student) (*Student, error) {
			return object, nil
		})

		assert.Error(t, err)
	})
}

func TestMongoRemove(t *testing.T) {
	t.Run("Remove", func(t *testing.T) {
		client, database := connectTestMongo(t)
		defer client.Disconnect(context.TODO())

		publisher := NewEventPublisher()
		repo, err := NewMongoRepository[*Student](database, &publisher)
		assert.NoError(t, err)
		object := Student{ID: uuid.New().String()}
		object.SetGrade("a")
		err = repo.Save(context.Background(), &object)
		assert.NoError(t, err)

		err = repo.Remove(context.Background(), object.ObjectID(), &object)

		assert.NoError(t, err)
		exists, err := repo.Exists(context.Background(), object.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		events := eventStreamFor(t, database, object.ID)
		assert.Len(t, events, 1)
		assert.Equal(t, REMOVED_EVENT_NAME, events[0].name)
	})
	t.Run("Remove unknown", func(t *testing.T) {
		client, database := connectTestMongo(t)
		defer client.Disconnect(context.TODO())

		publisher := NewEventPublisher()
		repo, err := NewMongoRepository[*Student](database, &publisher)
		assert.NoError(t, err)
		object := Student{ID: uuid.New().String()}
		object.SetGrade("a")
		err = repo.Save(context.Background(), &object)
		assert.NoError(t, err)

		err = repo.Remove(context.Background(), uuid.NewString(), &object)

		assert.Error(t, err)
		exists, err := repo.Exists(context.Background(), object.ID)
		assert.NoError(t, err)
		assert.True(t, exists)
	})
	t.Run("Remove is idempotent", func(t *testing.T) {
		client, database := connectTestMongo(t)
		defer client.Disconnect(context.TODO())

		publisher := NewEventPublisher()
		repo, err := NewMongoRepository[*Student](database, &publisher)
		assert.NoError(t, err)
		object := Student{ID: uuid.New().String()}
		object.SetGrade("a")
		err = repo.Save(context.Background(), &object)
		assert.NoError(t, err)

		err = repo.Remove(context.Background(), object.ObjectID(), &object)
		assert.NoError(t, err)

		err = repo.Remove(context.Background(), object.ObjectID(), &object)
		assert.NoError(t, err)
	})
}
