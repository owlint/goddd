package goddd

import (
	"context"
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
		repo := NewMongoRepository(database, &publisher)
		object := Student{ID: uuid.NewString()}

		object.SetGrade("a")
		repo.Save(&object)

		assert.Len(t, eventStreamFor(t, database, object.ObjectID()), 1)
	})
	t.Run("save is idempotent", func(t *testing.T) {
		client, database := connectTestMongo(t)
		defer client.Disconnect(context.TODO())

		publisher := NewEventPublisher()
		repo := NewMongoRepository(database, &publisher)
		object := Student{ID: uuid.NewString()}

		object.SetGrade("a")
		repo.Save(&object)
		assert.Len(t, eventStreamFor(t, database, object.ObjectID()), 1)

		object.SetGrade("a")
		repo.Save(&object)
		assert.Len(t, eventStreamFor(t, database, object.ObjectID()), 2)
		repo.Save(&object)
		assert.Len(t, eventStreamFor(t, database, object.ObjectID()), 2)
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
	repo := NewMongoRepository(database, &publisher)
	object := Student{}

	object.SetGrade("a")
	repo.Save(&object)

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
	repo := NewMongoRepository(database, &publisher)
	object := Student{}

	object.SetGrade("a")
	repo.Save(&object)
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
	repo := NewMongoRepository(database, &publisher)
	object := Student{ID: uuid.New().String()}

	object.SetGrade("a")
	object.SetGrade("a")
	repo.Save(&object)

	if len(eventStreamFor(t, database, object.ObjectID())) != 2 {
		t.Error("Should contain only one event")
		t.FailNow()
	}
}

func TestMongoSaveMultiplesObject(t *testing.T) {
	client, database := connectTestMongo(t)
	defer client.Disconnect(context.TODO())

	publisher := NewEventPublisher()
	repo := NewMongoRepository(database, &publisher)
	object := Student{ID: uuid.New().String()}
	object2 := Student{ID: uuid.New().String()}

	object.SetGrade("a")
	object.SetGrade("a")
	repo.Save(&object)

	object2.SetGrade("a")
	repo.Save(&object2)

	if len(eventStreamFor(t, database, object.ObjectID())) != 2 {
		t.Errorf("Should contain only one event : %d", len(eventStreamFor(t, database, object.ObjectID())))
		t.FailNow()
	}
}

func TestMongoLoad(t *testing.T) {
	client, database := connectTestMongo(t)
	defer client.Disconnect(context.TODO())

	publisher := NewEventPublisher()
	repo := NewMongoRepository(database, &publisher)
	object := Student{ID: uuid.New().String()}
	object2 := Student{ID: uuid.New().String()}

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

func TestMongoEventsSince(t *testing.T) {
	client, database := connectTestMongo(t)
	defer client.Disconnect(context.TODO())

	publisher := NewEventPublisher()
	repo := NewMongoRepository(database, &publisher)
	object := Student{ID: uuid.New().String()}
	object2 := Student{ID: uuid.New().String()}

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

func TestMongoEventsSinceLimit(t *testing.T) {
	client, database := connectTestMongo(t)
	defer client.Disconnect(context.TODO())

	publisher := NewEventPublisher()
	repo := NewMongoRepository(database, &publisher)
	object := Student{ID: uuid.New().String()}
	object2 := Student{ID: uuid.New().String()}

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

func TestMongoMementizerSnapshotSaved(t *testing.T) {
	client, database := connectTestMongo(t)
	defer client.Disconnect(context.TODO())

	publisher := NewEventPublisher()
	repo := NewMongoRepository(database, &publisher)
	object := StudentMemento{
		EventStream: &Stream{},
		ID:          uuid.New().String(),
	}

	for i := 0; i < 600; i++ {
		object.SetGrade(fmt.Sprintf("a%d", i))
	}
	repo.Save(&object)

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
	repo := NewMongoRepository(database, &publisher)
	object := Student{ID: uuid.New().String()}

	for i := 0; i < 600; i++ {
		object.SetGrade(fmt.Sprintf("a%d", i))
	}
	repo.Save(&object)

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
	repo := NewMongoRepository(database, &publisher)
	object := StudentMemento{
		EventStream: &Stream{},
		ID:          uuid.New().String(),
	}

	for i := 0; i < 600; i++ {
		object.SetGrade(fmt.Sprintf("a%d", i))
	}
	repo.Save(&object)
	for i := 0; i < 10; i++ {
		object.SetGrade(fmt.Sprintf("a%d", i))
	}
	repo.Save(&object)

	loadedObject := StudentMemento{
		EventStream: &Stream{},
	}
	err := repo.Load(object.ObjectID(), &loadedObject)

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
