package reflective

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	assert.Nil(t, nil)
}

func TestNoAnomaly(t *testing.T) {
	now := time.Now().UnixNano()

	RecordRead("key", "value", now)

	// If the write is after then it's never an anomaly
	assert.False(t, CheckWriteForAnomaly("key", "value2", now+1))
	assert.False(t, CheckWriteForAnomaly("key", "value", now+1))

	// If the write is before but the same value, it's not an anomaly
	assert.False(t, CheckWriteForAnomaly("key", "value", now-1))

	// If the write is at the same time then it's never an anomaly
	assert.False(t, CheckWriteForAnomaly("key", "value", now))
	assert.False(t, CheckWriteForAnomaly("key", "value2", now))

	// If the write is for a key we've never read then it's never an anomaly
	assert.False(t, CheckWriteForAnomaly("key2", "value", now))
	assert.False(t, CheckWriteForAnomaly("key2", "value", now-1))
	assert.False(t, CheckWriteForAnomaly("key2", "value", now+1))
	assert.False(t, CheckWriteForAnomaly("key2", "value2", now-1))
}

func TestAnomaly(t *testing.T) {
	now := time.Now().UnixNano()

	RecordRead("key", "value", now)

	// If the write is before and a different value, it's an anomaly
	assert.True(t, CheckWriteForAnomaly("key", "value2", now-1))
	assert.True(t, CheckWriteForAnomaly("key", "value3", now-2))
}
