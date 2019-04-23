package reflective

import (
	"bytes"
	"time"
)

// readRecord stores when and what was read
type readRecord struct {
	whenReadNano int64 // use time.Now().UnixNano()
	valueRead    []byte
	// TODO maybe keep how many times read?
	// and latest value read before seeing a write?
	// if so don't need an array as mentioned
	// in the TODO below
}

// read stores a read record for each key
// TODO need array?
var keyReadRecords map[string]readRecord = map[string]readRecord{}

// keyAnomalies stores a number of anomalies
// for each key
var keyAnomalies map[string]int = map[string]int{}

// RecordRead records a read for key which returned value at time when
func RecordRead(key string, value []byte, when int64) {
	keyReadRecords[key] = readRecord{whenReadNano: when, valueRead: value}
}

// RecordRead records a read for key which returned value at NOW
func RecordReadNow(key string, value []byte) {
	RecordRead(key, value, time.Now().UnixNano())
}

// CheckWriteForAnomaly checks a write of value to key at a time when
func CheckWriteForAnomaly(key string, value []byte, when int64) bool {
	readRecord, found := keyReadRecords[key]
	if !found {
		return false
	}

	// If the write is in the past and it's different than
	// what was returned on the read then it's an anomaly
	if readRecord.whenReadNano > when {
		if !bytes.Equal(readRecord.valueRead, value) {
			registerAnomaly(key)
			return true
		}
	}

	// TODO if not an anomaly, do we remove the read record?

	return false
}

// CheckWriteForAnomaly checks a write of value to key at NOW
func CheckWriteForAnomalyNow(key string, value []byte) bool {
	return CheckWriteForAnomaly(key, value, time.Now().UnixNano())
}

// registerAnomaly registers an anomaly for key
func registerAnomaly(key string) {
	val, found := keyAnomalies[key]
	if !found {
		val = 0
	}
	val++
	keyAnomalies[key] = val
}
