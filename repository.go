package goddd

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
)

var ConcurrencyError = errors.New("concurrency error while saving")
var InvalidUpdateCallback = errors.New("callback should return a non nil object if error is nil")

type Repository[T DomainObject] interface {
	Save(ctx context.Context, object T) error
	Load(ctx context.Context, objectID string, object T) error
	Exists(ctx context.Context, objectID string) (bool, error)
	EventsSince(ctx context.Context, time time.Time, limit int) ([]Event, error)
	Update(ctx context.Context, objectID string, object T, nbRetries int, updater func(T) (T, error)) (T, error)
	Remove(ctx context.Context, objectID string, object T) error
}

func unsavedEvents(objectEvents []Event, knownEventIDs []string) []Event {
	knownIDs := make(map[string]struct{}, len(knownEventIDs))
	events := make([]Event, 0)

	for _, eventID := range knownEventIDs {
		knownIDs[eventID] = struct{}{}
	}

	for _, event := range objectEvents {
		if _, exists := knownIDs[event.Id()]; !exists {
			events = append(events, event)
		}
	}

	return events
}

func NewIdentity(objectType string) string {
	return fmt.Sprintf("%s-%s", objectType, uuid.New().String())
}

func Encode(object interface{}) ([]byte, error) {
	var data bytes.Buffer
	encoder := gob.NewEncoder(&data)
	err := encoder.Encode(object)
	if err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}

func Decode(object interface{}, data []byte) error {
	encoder := gob.NewDecoder(bytes.NewBuffer(data))
	err := encoder.Decode(object)
	if err != nil {
		return err
	}
	return nil
}

func repoUpdate[T DomainObject](ctx context.Context, repo Repository[T], objectID string, object T, nbRetries int, updater func(T) (T, error)) (T, error) {
	if nbRetries < 0 {
		return object, errors.New("negative number of retries")
	}

	var err error
	for i := 0; i <= nbRetries; i++ {
		err = repo.Load(ctx, objectID, object)
		if err != nil {
			return object, err
		}
		object, err = updater(object)
		if err != nil {
			return object, err
		}
		if reflect.ValueOf(object).IsNil() {
			return object, InvalidUpdateCallback
		}
		err = repo.Save(ctx, object)
		if errors.Is(err, ConcurrencyError) {
			continue
		}
		return object, err
	}
	return object, err
}
