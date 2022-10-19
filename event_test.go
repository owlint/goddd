package goddd

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEventCreation(t *testing.T) {
	objectID := uuid.New().String()
	before := time.Now().UnixNano()
	time.Sleep(5 * time.Millisecond)
	event := NewEvent(objectID, "eventCreated", 3, []byte{1, 2, 3})
	time.Sleep(5 * time.Millisecond)
	now := time.Now().UnixNano()

	assert.Equal(t, "eventCreated", event.Name())
	assert.Equal(t, 3, event.Version())
	assert.True(t, strings.HasPrefix(event.Id(), objectID))
	assert.True(t, len(event.Id()) > len(objectID))
	assert.True(t, event.Timestamp() >= before && event.Timestamp() <= now)
	assert.Equal(t, objectID, event.ObjectId())
	assert.Equal(t, []byte{1, 2, 3}, event.Payload())
}

func benchmarkEventCreation(objectID string, eventName string, version int, payload []byte, b *testing.B) {
	for n := 0; n < b.N; n++ {
		NewEvent(objectID, eventName, version, payload)
	}
}

func BenchmarkEventCreation(b *testing.B) {
	objectID := uuid.New().String()
	benchmarkEventCreation(objectID, "eventName", 3, []byte{1, 2, 3}, b)
}

func TestSerialize(t *testing.T) {
	event := NewEvent(uuid.New().String(), "eventCreated", 3, []byte{1, 2, 3})
	serialized, err := event.Serialize()

	assert.NoError(t, err)

	reloaded, err := Deserialize(serialized)
	assert.NoError(t, err)

	assert.Equal(t, event.ObjectId(), reloaded.ObjectId())
	assert.InDelta(t, event.Timestamp(), reloaded.Timestamp(), float64(time.Second))
	assert.Equal(t, event.Name(), reloaded.Name())
	assert.Equal(t, event.Version(), reloaded.Version())
	assert.Equal(t, event.Payload(), reloaded.Payload())
}
