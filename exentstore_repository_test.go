package goddd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func exentStreamFor(t *testing.T, objectID string) []exentstoreStreamEvent {
	resp, err := http.Get(fmt.Sprintf("http://localhost:4000/api/streams/%s?after_stream_version=0", objectID))

	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Error(resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	stream := exentstoreStream{}
	err = json.Unmarshal(body, &stream)
	if err != nil {
		t.Error(err)
	}
	return stream.Events
}

func TestExentStoreSave(t *testing.T) {
	publisher := NewEventPublisher()
	repo := NewExentStoreRepository("http://localhost:4000", &publisher)
	object := testDomainObject{ID: NewIdentity("TestDomainObject")}

	object.Method(5)
	err := repo.Save(&object)

	assert.Nil(t, err)
	assert.Len(t, exentStreamFor(t, object.ObjectID()), 1)
}

func TestExentStorePublished(t *testing.T) {
	receiver := testReceiver{
		events: make([]Event, 0),
	}
	publisher := NewEventPublisher()
	publisher.Register(&receiver)
	publisher.Wait = true
	repo := NewExentStoreRepository("http://localhost:4000", &publisher)
	object := testDomainObject{ID: NewIdentity("TestDomainObject")}

	object.Method(5)
	repo.Save(&object)

	if len(receiver.events) != 1 {
		t.Error("No events received")
		t.FailNow()
	}
}

func TestExentStorePublishedNoWait(t *testing.T) {
	receiver := testReceiver{
		events: make([]Event, 0),
	}
	publisher := NewEventPublisher()
	publisher.Register(&receiver)
	repo := NewExentStoreRepository("http://localhost:4000", &publisher)
	object := testDomainObject{ID: NewIdentity("TestDomainObject")}

	object.Method(5)
	repo.Save(&object)
	time.Sleep(500 * time.Millisecond)

	if len(receiver.events) != 1 {
		t.Error("No events received")
		t.FailNow()
	}
}
func TestExentStoreSaveVersionCorrection(t *testing.T) {
	publisher := NewEventPublisher()
	repo := NewExentStoreRepository("http://localhost:4000", &publisher)
	object := testDomainObject{ID: NewIdentity("TestDomainObject")}

	object.Method(5)
	object.Method(7)
	repo.Save(&object)

	object.Method(7)
	object.Stream.events[len(object.Stream.events)-1].version = 5
	repo.Save(&object)

	if len(exentStreamFor(t, object.ObjectID())) != 3 {
		t.Error("Should contain only one event")
		t.FailNow()
	}
}

func TestExentStoreSaveMultiples(t *testing.T) {
	publisher := NewEventPublisher()
	repo := NewExentStoreRepository("http://localhost:4000", &publisher)
	object := testDomainObject{ID: NewIdentity("TestDomainObject")}

	object.Method(5)
	object.Method(7)
	repo.Save(&object)

	if len(exentStreamFor(t, object.ObjectID())) != 2 {
		t.Error("Should contain only one event")
		t.FailNow()
	}
}

func TestExentStoreSaveMultiplesObject(t *testing.T) {
	publisher := NewEventPublisher()
	repo := NewExentStoreRepository("http://localhost:4000", &publisher)
	object := testDomainObject{ID: NewIdentity("TestDomainObject")}
	object2 := testDomainObject{ID: NewIdentity("TestDomainObject")}

	object.Method(5)
	object.Method(7)
	repo.Save(&object)

	object2.Method(10)
	repo.Save(&object2)

	if len(exentStreamFor(t, object.ObjectID())) != 2 {
		t.Errorf("Should contain only one event : %d", len(exentStreamFor(t, object.ObjectID())))
		t.FailNow()
	}
}

func TestExentStoreLoad(t *testing.T) {
	publisher := NewEventPublisher()
	repo := NewExentStoreRepository("http://localhost:4000", &publisher)
	object := testDomainObject{ID: NewIdentity("TestDomainObject")}
	object2 := testDomainObject{ID: NewIdentity("TestDomainObject")}

	object.Method(5)
	object.Method(7)
	repo.Save(&object)

	object2.Method(10)
	repo.Save(&object2)

	loadedObject := testDomainObject{}
	err := repo.Load(object.ObjectID(), &loadedObject)

	if err != nil {
		t.Errorf("No error should occur on load : %v", err)
		t.FailNow()
	}

	if loadedObject.number != 12 {
		t.Error("Wrong value")
		t.FailNow()
	}
}
