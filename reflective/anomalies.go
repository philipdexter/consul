package reflective

import (
	"time"
)

// readRecord stores when and what was read
type readRecord struct {
	whenReadNano int64 // use time.Now().UnixNano()
	valueRead    string
}

// read stores a read record for each key
var keyReadRecords map[string]readRecord = map[string]readRecord{}

// RecordRead records a read for key which returned value at time when
func RecordRead(key, value string, when int64) {
	keyReadRecords[key] = readRecord{whenReadNano: when, valueRead: value}
}

// RecordRead records a read for key which returned value at NOW
func RecordReadNow(key, value string) {
	RecordRead(key, value, time.Now().UnixNano())
}

// CheckWriteForAnomaly checks a write of value to key at a time when
func CheckWriteForAnomaly(key, value string, when int64) bool {
	readRecord, found := keyReadRecords[key]
	if !found {
		return false
	}

	// If the write is in the past and it's different than
	// what was returned on the read then it's an anomaly
	if readRecord.whenReadNano > when {
		if readRecord.valueRead != value {
			return true
		}
	}

	return false
}

// CheckWriteForAnomaly checks a write of value to key at NOW
func CheckWriteForAnomalyNow(key, value string) bool {
	return CheckWriteForAnomaly(key, value, time.Now().UnixNano())
}
