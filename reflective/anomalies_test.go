package reflective

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func resetLogic() {
	keyReadRecords = map[string]readRecord{}
	keyAnomalies = map[string]int{}
}

func TestNoAnomaly(t *testing.T) {
	resetLogic()

	now := time.Now().UnixNano()

	RecordRead("key", []byte("value"), now)

	// If the write is after then it's never an anomaly
	assert.False(t, CheckWriteForAnomaly("key", []byte("value2"), now+1))
	assert.False(t, CheckWriteForAnomaly("key", []byte("value"), now+1))

	// If the write is before but the same value, it's not an anomaly
	assert.False(t, CheckWriteForAnomaly("key", []byte("value"), now-1))

	// If the write is at the same time then it's never an anomaly
	assert.False(t, CheckWriteForAnomaly("key", []byte("value"), now))
	assert.False(t, CheckWriteForAnomaly("key", []byte("value2"), now))

	// If the write is for a key we've never read then it's never an anomaly
	assert.False(t, CheckWriteForAnomaly("key2", []byte("value"), now))
	assert.False(t, CheckWriteForAnomaly("key2", []byte("value"), now-1))
	assert.False(t, CheckWriteForAnomaly("key2", []byte("value"), now+1))
	assert.False(t, CheckWriteForAnomaly("key2", []byte("value2"), now-1))

	anoms, _ := keyAnomalies["key"]
	assert.Equal(t, 0, anoms)
	anoms, _ = keyAnomalies["key2"]
	assert.Equal(t, 0, anoms)
}

func TestAnomaly(t *testing.T) {
	resetLogic()

	now := time.Now().UnixNano()

	RecordRead("key", []byte("value"), now)

	// If the write is before and a different value, it's an anomaly
	assert.True(t, CheckWriteForAnomaly("key", []byte("value2"), now-1))
	assert.True(t, CheckWriteForAnomaly("key", []byte("value3"), now-2))

	anoms, _ := keyAnomalies["key"]
	assert.Equal(t, 2, anoms)
}
