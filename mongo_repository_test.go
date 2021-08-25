package goddd

import (
	"context"
	"testing"
	"time"

    "github.com/google/uuid"

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
	client, database := connectTestMongo(t)
	defer client.Disconnect(context.TODO())

	publisher := NewEventPublisher()
	repo := NewMongoRepository(database, &publisher)
	object := Student{}

	object.SetGrade("a")
	repo.Save(&object)

	if len(eventStreamFor(t, database, object.ObjectID())) != 1 {
		t.Error("Should contain only one event")
		t.FailNow()
	}
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
