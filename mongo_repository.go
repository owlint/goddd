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

func (r *MongoRepository[T]) Save(ctx context.Context, object T) error {
	events := object.CollectUnsavedEvents()

	records := toRecords(events)
	_, err := r.collection.InsertMany(ctx, records)
	if err != nil && !errors.Is(err, mongo.ErrEmptySlice) {
		if mongo.IsDuplicateKeyError(err) {
			return ConcurrencyError
		}
		return err
	}

	r.publisher.Publish(events)
	return r.saveSnapshot(ctx, object)
}

func (r *MongoRepository[T]) Update(ctx context.Context, objectID string, object T, nbRetries int, updater func(T) (T, error)) (T, error) {
	return repoUpdate[T](ctx, r, objectID, object, nbRetries, updater)
}

func (r *MongoRepository[T]) saveSnapshot(ctx context.Context, object T) error {
	var objectInter interface{} = object
	mementizer, isMemento := objectInter.(DomainObjectMemento)

	if isMemento {
		lastSnapshot, err := r.lastSnapshot(ctx, object.ObjectID())
		if err != nil {
			return fmt.Errorf("Could not save snapshot : %s", err.Error())
		}
		lastVersion := 0
		if lastSnapshot != nil {
			lastVersion = lastSnapshot.Version
		}
		if object.LastVersion()-lastVersion > 500 {
			err := r.persistSnapshot(ctx, object, mementizer)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *MongoRepository[T]) persistSnapshot(ctx context.Context, object T, mementizer DomainObjectMemento) error {
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

	_, err = r.snapshotsCollection.UpdateOne(ctx, filter, update, options)

	if r.snapshotsCache != nil {
		r.snapshotsCache.Set(object.ObjectID(), snap, 1)
	}

	return err
}

func (r *MongoRepository[T]) Load(ctx context.Context, objectID string, object T) error {
	exist, err := r.Exists(ctx, objectID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("Cannot load unknown object")
	}

	object.Clear()

	snapshot, err := r.lastSnapshot(ctx, objectID)
	if err != nil {
		return err
	}

	var objectEvents []Event
	if snapshot != nil {
		err = r.reloadSnapshot(snapshot, object)
		if err != nil {
			return err
		}
		objectEvents, err = r.ObjectEventsSinceVersion(ctx, objectID, snapshot.Version)
	} else {
		objectEvents, err = r.objectRepositoryEvents(ctx, objectID)
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

func (r *MongoRepository[T]) Exists(ctx context.Context, objectId string) (bool, error) {
	filter := bson.D{{"objectid", objectId}, {"name", bson.D{{"$ne", REMOVED_EVENT_NAME}}}}
	result := r.collection.FindOne(ctx, filter)
	if result != nil && result.Err() == mongo.ErrNoDocuments {
		return false, nil
	} else if result.Err() != nil {
		return false, result.Err()
	}
	return true, nil
}

func (r *MongoRepository[T]) EventsSince(ctx context.Context, timestamp time.Time, limit int) ([]Event, error) {
	records := make([]record, 0)

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"timestamp": 1})
	findOptions.SetLimit(int64(limit))
	filter := bson.M{"timestamp": bson.M{
		"$gte": timestamp.UnixNano(),
	}}
	listCursor, err := r.collection.Find(ctx, filter, findOptions)
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

func (r *MongoRepository[T]) ObjectEventsSinceVersion(ctx context.Context, objectID string, version int) ([]Event, error) {
	records := make([]record, 0)

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"version": 1})
	filter := bson.M{
		"version": bson.M{
			"$gt": version,
		},
		"objectid": objectID,
	}
	listCursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer listCursor.Close(ctx)

	err = listCursor.All(ctx, &records)
	if err != nil {
		return fromRecords(records), err
	}

	return fromRecords(records), nil
}

func (r *MongoRepository[T]) objectRepositoryEvents(ctx context.Context, objectID string) ([]Event, error) {
	records := make([]record, 0)

	filter := bson.D{{"objectid", objectID}}
	opts := options.Find().SetSort(bson.M{"version": 1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return fromRecords(records), err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &records)
	if err != nil {
		return fromRecords(records), err
	}

	return fromRecords(records), nil
}

func (r *MongoRepository[T]) lastSnapshot(ctx context.Context, objectID string) (*snapshot, error) {
	if r.snapshotsCache != nil {
		snap, ok := r.snapshotsCache.Get(objectID)
		if ok {
			snapshot := snap.(snapshot)
			return &snapshot, nil
		}
	}

	filter := bson.D{{"objectid", objectID}}
	result := r.snapshotsCollection.FindOne(ctx, filter)
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

func (r *MongoRepository[T]) lastVersion(ctx context.Context, objectID string) (int64, error) {
	cursor, err := r.collection.Aggregate(
		ctx,
		bson.A{
			bson.D{{"$match", bson.D{{"objectid", objectID}}}},
			bson.D{
				{"$group",
					bson.D{
						{"_id", "$objectid"},
						{"lastVersion", bson.D{{"$max", "$version"}}},
					},
				},
			},
		},
	)
	if err != nil {
		return -1, err
	}
	result := make([]struct {
		ObjectID    string `bson:"_id"`
		LastVersion int64  `bson:"lastVersion"`
	}, 0)
	err = cursor.All(ctx, &result)
	if err != nil {
		return -1, err
	}
	if len(result) == 0 {
		return -1, errors.New("unknown object")
	}
	return result[0].LastVersion, nil
}

func (r *MongoRepository[T]) alreadyRemoved(ctx context.Context, objectID string) (bool, error) {
	filter := bson.D{{"objectid", objectID}, {"name", bson.D{{"$eq", REMOVED_EVENT_NAME}}}}
	result := r.collection.FindOne(ctx, filter)
	if result != nil && errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return false, nil
	} else if result.Err() != nil {
		return false, result.Err()
	}
	return true, nil
}

func (r *MongoRepository[T]) Remove(ctx context.Context, objectID string, object T) error {
	if alreadyRemoved, _ := r.alreadyRemoved(ctx, objectID); alreadyRemoved {
		return nil
	}
	if exists, err := r.Exists(ctx, objectID); err != nil || !exists {
		return errors.New("cannot load unknown object")
	}

	lastVersion, err := r.lastVersion(ctx, objectID)
	if err != nil {
		return err
	}
	event := NewEvent(objectID, REMOVED_EVENT_NAME, int(lastVersion)+1, []byte{})
	events := []Event{event}
	records := toRecords(events)
	_, err = r.collection.InsertMany(ctx, records)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ConcurrencyError
		}
		return err
	}

	_, err = r.collection.DeleteMany(
		ctx,
		bson.D{
			bson.E{"objectid", objectID},
			bson.E{"name", bson.D{{"$ne", REMOVED_EVENT_NAME}}},
		},
	)
	if err != nil {
		return err
	}

	_, err = r.snapshotsCollection.DeleteMany(ctx, bson.D{{"objectid", object}})
	if err != nil {
		return err
	}

	r.publisher.Publish(events)

	return nil
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
	collection := mongoDB.Collection("event_store")
	_, err := collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				bson.E{
					Key:   "objectid",
					Value: 1,
				},
			},
			Options: options.Index().SetName("objectID").SetUnique(false).SetBackground(true),
		},
		{
			Keys: bson.D{
				bson.E{
					Key:   "objectid",
					Value: 1,
				},
				bson.E{
					Key:   "version",
					Value: 1,
				},
			},
			Options: options.Index().SetName("objectID_version_unique").SetUnique(true).SetBackground(true),
		},
		{
			Keys: bson.D{
				bson.E{
					Key:   "id",
					Value: 1,
				},
			},
			Options: options.Index().SetName("eventid_unique").SetUnique(true).SetBackground(true),
		},
		{
			Keys: bson.D{
				bson.E{
					Key:   "timestamp",
					Value: 1,
				},
			},
			Options: options.Index().SetName("timestamp_index").SetUnique(false).SetBackground(true),
		},
		{
			Keys: bson.D{
				bson.E{
					Key:   "objectid",
					Value: 1,
				},
				bson.E{
					Key:   "name",
					Value: 1,
				},
			},
			Options: options.Index().SetName("objectID_name").SetUnique(false).SetBackground(true),
		},
	})

	return err
}
