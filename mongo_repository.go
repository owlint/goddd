package goddd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type snapshot struct {
	ObjectID string
	Version  int
	Payload  []byte
}

type record struct {
	ID        string
	Version   int
	ObjectID  string
	Timestamp int64
	Name      string
	Payload   []byte
}

type MongoRepository struct {
	collection          *mongo.Collection
	snapshotsCollection *mongo.Collection
	publisher           *EventPublisher
	snapshotsCache      *ristretto.Cache
}

func NewMongoRepository(database *mongo.Database, publisher *EventPublisher) MongoRepository {
	cache, _ := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e3,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	return MongoRepository{
		collection:          database.Collection("event_store"),
		snapshotsCollection: database.Collection("domain_event_snapshots"),
		publisher:           publisher,
		snapshotsCache:      cache,
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

	err = r.saveSnapshot(object)

	return err
}

func (r *MongoRepository) saveSnapshot(object DomainObject) error {
	var objectInter interface{} = object
	mementizer, isMemento := objectInter.(DomainObjectMemento)

	if isMemento {
		lastSnapshot, err := r.lastSnapshot(object.ObjectID())
		if err != nil {
			return fmt.Errorf("Could not save snapshot : %s", err.Error())
		}
		lastVersion := 0
		if lastSnapshot != nil {
			lastVersion = lastSnapshot.Version
		}
		if object.LastVersion()-lastVersion > 500 {
			err := r.persistSnapshot(object, mementizer)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *MongoRepository) persistSnapshot(object DomainObject, mementizer DomainObjectMemento) error {
	memento, err := mementizer.DumpMemento()
	if err != nil {
		return err
	}
	bytePayload, err := memento.MarshalMsg(nil)
	if err != nil {
		return err
	}

	snap := snapshot{
		ObjectID: object.ObjectID(),
		Version:  object.LastVersion(),
		Payload:  bytePayload,
	}

	update := bson.M{
		"$set": snap,
	}
	options := options.Update().SetUpsert(true)
	filter := bson.M{
		"objectid": object.ObjectID(),
	}

	_, err = r.snapshotsCollection.UpdateOne(context.Background(), filter, update, options)

	if r.snapshotsCache != nil {
		r.snapshotsCache.Set(object.ObjectID(), snap, 1)
	}

	return err
}

func (r *MongoRepository) Load(objectID string, object DomainObject) error {
	exist, err := r.Exists(objectID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("Cannot load unknown object")
	}

	object.Clear()

	snapshot, err := r.lastSnapshot(objectID)
	if err != nil {
		return err
	}

	var objectEvents []Event
	if snapshot != nil {
		err = r.reloadSnapshot(snapshot, object)
		if err != nil {
			return err
		}
		objectEvents, err = r.ObjectEventsSinceVersion(objectID, snapshot.Version)
	} else {
		objectEvents, err = r.objectRepositoryEvents(objectID)
	}

	if err != nil {
		return err
	}

	for _, event := range objectEvents {
		err = object.LoadEvent(object, event)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *MongoRepository) reloadSnapshot(snapshot *snapshot, object DomainObject) error {
	var objectInter interface{} = object
	mementizer, isMemento := objectInter.(DomainObjectMemento)

	if isMemento {
		mementizer.SetVersion(snapshot.Version)
		return mementizer.ApplyMemento(snapshot.Payload)
	}
	return nil
}

func (r *MongoRepository) Exists(objectId string) (bool, error) {
	filter := bson.D{{"objectid", objectId}}
	result := r.collection.FindOne(context.TODO(), filter)
	if result != nil && result.Err() == mongo.ErrNoDocuments {
		return false, nil
	} else if result.Err() != nil {
		return false, result.Err()
	}
	return true, nil
}

func (r *MongoRepository) EventsSince(timestamp time.Time, limit int) ([]Event, error) {
	records := make([]record, 0)

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"timestamp": 1})
	findOptions.SetLimit(int64(limit))
	filter := bson.M{"timestamp": bson.M{
		"$gte": timestamp.UnixNano(),
	}}
	listCursor, err := r.collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		return nil, err
	}

	err = listCursor.All(context.TODO(), &records)
	if err != nil {
		return fromRecords(records), err
	}

	return fromRecords(records), nil
}

func (r *MongoRepository) ObjectEventsSinceVersion(objectID string, version int) ([]Event, error) {
	records := make([]record, 0)

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"version": 1})
	filter := bson.M{
		"version": bson.M{
			"$gt": version,
		},
		"objectid": objectID,
	}
	listCursor, err := r.collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		return nil, err
	}

	err = listCursor.All(context.TODO(), &records)
	if err != nil {
		return fromRecords(records), err
	}

	return fromRecords(records), nil
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

func (r *MongoRepository) lastSnapshot(objectID string) (*snapshot, error) {
	if r.snapshotsCache != nil {
		snap, ok := r.snapshotsCache.Get(objectID)
		if ok {
			snapshot := snap.(snapshot)
			return &snapshot, nil
		}
	}

	filter := bson.D{{"objectid", objectID}}
	result := r.snapshotsCollection.FindOne(context.TODO(), filter)
	if result != nil && result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	} else if result.Err() != nil {
		return nil, result.Err()
	}

	lastSnapshot := snapshot{}
	err := result.Decode(&lastSnapshot)
	if err != nil {
		return nil, err
	}

	return &lastSnapshot, nil
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
			objectID:  record.ObjectID,
			timestamp: record.Timestamp,
			name:      record.Name,
			payload:   record.Payload,
		}
	}

	return events
}
