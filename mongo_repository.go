package goddd

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type record struct {
	ID        string
	Version   int
	ObjectID  string
	Timestamp int64
	Name      string
	Payload   interface{}
}

type MongoRepository struct {
	collection *mongo.Collection
	publisher  EventPublisher
}

func NewMongoRepository(database *mongo.Database, publisher EventPublisher) MongoRepository {
	return MongoRepository{
		collection: database.Collection("event_store"),
		publisher:  publisher,
	}
}

func (r *MongoRepository) Save(object DomainObject) error {
	objectRepoEvents, err := r.objectRepositoryEvents(object.ObjectID())

	if err != nil {
		return err
	}

	objectEvents := object.Events()
	eventToAdd := unsavedEvents(objectEvents, objectRepoEvents)

	records := toRecords(eventToAdd)
	r.collection.InsertMany(context.TODO(), records)

	r.publisher.Publish(eventToAdd)

	return nil
}

func (r *MongoRepository) Load(object DomainObject) error {
	objectID := object.ObjectID()

	exist, err := r.Exists(objectID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("Cannot load unknown object")
	}

	objectEvents, err := r.objectRepositoryEvents(objectID)
	if err != nil {
		return err
	}

	for _, event := range objectEvents {
		object.AddEvent(object, event.Name(), event.Payload())
	}

	return nil
}

func (r *MongoRepository) Exists(objectId string) (bool, error) {
	filter := bson.D{{"objectid", objectId}}
	count, err := r.collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *MongoRepository) objectRepositoryEvents(objectID string) ([]Event, error) {
	records := make([]record, 0)

	filter := bson.D{{"objectid", objectID}}
	cursor, err := r.collection.Find(context.TODO(), filter)

	if err != nil {
		return fromRecords(records), err
	}

	err = cursor.All(context.TODO(), &records)
	if err != nil {
		return fromRecords(records), err
	}

	return fromRecords(records), nil
}

func toRecords(events []Event) []interface{} {
	records := make([]interface{}, len(events))
	for i, event := range events {
		records[i] = record{
			ID:        event.Id(),
			Version:   event.Version(),
			ObjectID:  event.ObjectId(),
			Timestamp: event.Timestamp(),
			Name:      event.Name(),
			Payload:   event.Payload(),
		}
	}
	return records
}

func fromRecords(records []record) []Event {
	events := make([]Event, len(records))
	for i, record := range records {
		events[i] = Event{
			id:        record.ID,
			version:   record.Version,
			objectId:  record.ObjectID,
			timestamp: record.Timestamp,
			name:      record.Name,
			payload:   record.Payload,
		}
	}

	return events
}
