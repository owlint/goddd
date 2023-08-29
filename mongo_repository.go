package goddd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/golang-migrate/migrate/v4"
	mongomigrate "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

type MongoRepository[T DomainObject] struct {
	collection          *mongo.Collection
	snapshotsCollection *mongo.Collection
	publisher           *EventPublisher
	snapshotsCache      *ristretto.Cache
}

func NewMongoRepository[T DomainObject](database *mongo.Database, publisher *EventPublisher) (*MongoRepository[T], error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e3,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}
	err = MigrateMongoDB(database, "./migrations")
	if err != nil {
		return nil, err
	}
	return &MongoRepository[T]{
		collection:          database.Collection("event_store"),
		snapshotsCollection: database.Collection("domain_event_snapshots"),
		publisher:           publisher,
		snapshotsCache:      cache,
	}, nil
}

func (r *MongoRepository[T]) Save(object T) error {
	events := object.CollectUnsavedEvents()

	records := toRecords(events)
	_, err := r.collection.InsertMany(context.Background(), records)
	if err != nil && !errors.Is(err, mongo.ErrEmptySlice) {
		if mongo.IsDuplicateKeyError(err) {
			return ConcurrencyError
		}
		return err
	}

	r.publisher.Publish(events)
	err = r.saveSnapshot(object)

	return err
}

func (r *MongoRepository[T]) Update(objectID string, object T, nbRetries int, updater func(T) (T, error)) (T, error) {
	return repoUpdate[T](r, objectID, object, nbRetries, updater)
}

func (r *MongoRepository[T]) saveSnapshot(object T) error {
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

func (r *MongoRepository[T]) persistSnapshot(object T, mementizer DomainObjectMemento) error {
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

func (r *MongoRepository[T]) Load(objectID string, object T) error {
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

func (r *MongoRepository[T]) reloadSnapshot(snapshot *snapshot, object T) error {
	var objectInter interface{} = object
	mementizer, isMemento := objectInter.(DomainObjectMemento)

	if isMemento {
		mementizer.SetVersion(snapshot.Version)
		return mementizer.ApplyMemento(snapshot.Payload)
	}
	return nil
}

func (r *MongoRepository[T]) Exists(objectId string) (bool, error) {
	filter := bson.D{{"objectid", objectId}}
	result := r.collection.FindOne(context.Background(), filter)
	if result != nil && result.Err() == mongo.ErrNoDocuments {
		return false, nil
	} else if result.Err() != nil {
		return false, result.Err()
	}
	return true, nil
}

func (r *MongoRepository[T]) EventsSince(timestamp time.Time, limit int) ([]Event, error) {
	records := make([]record, 0)

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"timestamp": 1})
	findOptions.SetLimit(int64(limit))
	filter := bson.M{"timestamp": bson.M{
		"$gte": timestamp.UnixNano(),
	}}
	listCursor, err := r.collection.Find(context.Background(), filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer listCursor.Close(context.Background())

	err = listCursor.All(context.Background(), &records)
	if err != nil {
		return fromRecords(records), err
	}

	return fromRecords(records), nil
}

func (r *MongoRepository[T]) ObjectEventsSinceVersion(objectID string, version int) ([]Event, error) {
	records := make([]record, 0)

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"version": 1})
	filter := bson.M{
		"version": bson.M{
			"$gt": version,
		},
		"objectid": objectID,
	}
	listCursor, err := r.collection.Find(context.Background(), filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer listCursor.Close(context.Background())

	err = listCursor.All(context.Background(), &records)
	if err != nil {
		return fromRecords(records), err
	}

	return fromRecords(records), nil
}

func (r *MongoRepository[T]) objectRepositoryEvents(objectID string) ([]Event, error) {
	records := make([]record, 0)

	filter := bson.D{{"objectid", objectID}}
	opts := options.Find().SetSort(bson.M{"version": 1})

	cursor, err := r.collection.Find(context.Background(), filter, opts)
	if err != nil {
		return fromRecords(records), err
	}
	defer cursor.Close(context.Background())

	err = cursor.All(context.Background(), &records)
	if err != nil {
		return fromRecords(records), err
	}

	return fromRecords(records), nil
}

func (r *MongoRepository[T]) lastSnapshot(objectID string) (*snapshot, error) {
	if r.snapshotsCache != nil {
		snap, ok := r.snapshotsCache.Get(objectID)
		if ok {
			snapshot := snap.(snapshot)
			return &snapshot, nil
		}
	}

	filter := bson.D{{"objectid", objectID}}
	result := r.snapshotsCollection.FindOne(context.Background(), filter)
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
func MigrateMongoDB(mongoDB *mongo.Database, dir string) error {
	driver, err := mongomigrate.WithInstance(mongoDB.Client(), &mongomigrate.Config{
		DatabaseName:         mongoDB.Name(),
		MigrationsCollection: "goddd_migrations",
		TransactionMode:      false,
		Locking: mongomigrate.Locking{
			CollectionName: "goddd_migration_locking",
			Timeout:        5,
			Enabled:        true,
			Interval:       1,
		},
	})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+dir,
		"mongodb",
		driver,
	)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
