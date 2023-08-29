package goddd

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ConcurrencyError = errors.New("Concurrency error while saving")

type Repository[T DomainObject] interface {
	Save(object T) error
	Load(objectID string, object T) error
	Exists(objectID string) (bool, error)
	EventsSince(time time.Time, limit int) ([]Event, error)
	Update(objectID string, object T, nbRetries int, updater func(T) (T, error)) (T, error)
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

func repoUpdate[T DomainObject](repo Repository[T], objectID string, object T, nbRetries int, updater func(T) (T, error)) (T, error) {
	if nbRetries < 0 {
		return object, errors.New("negative number of retries")
	}

	var err error
	for i := 0; i <= nbRetries; i++ {
		err = repo.Load(objectID, object)
		if err != nil {
			return object, err
		}
		object, err = updater(object)
		if err != nil {
			return object, err
		}
		err = repo.Save(object)
		if errors.Is(err, ConcurrencyError) {
			continue
		}
		return object, err
	}
	return object, err
}
