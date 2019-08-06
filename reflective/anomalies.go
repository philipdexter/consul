package reflective

import (
	"bytes"
	"sync"
	"time"
)

// readRecord stores when and what was read
type readRecord struct {
	whenReadNano int64 // use time.Now().UnixNano()
	valueRead    []byte
	// TODO maybe keep how many times read?
}

// read stores a read record for each key
var keyReadRecords map[string]readRecord = map[string]readRecord{}
var keyReadRecordsLock = sync.Mutex{}

// keyAnomalies stores a number of anomalies
// for each key
var keyAnomalies map[string]int = map[string]int{}
var keyAnomaliesLock = sync.Mutex{}

// keyReads stores the total number of reads
// performed for each key
var keyReads map[string]int = map[string]int{}
var keyReadsLock = sync.Mutex{}

// RecordRead records a read for key which returned value at time when
func RecordRead(key string, value []byte, when int64) {
	keyReadRecordsLock.Lock()
	keyReadRecords[key] = readRecord{whenReadNano: when, valueRead: value}
	keyReadRecordsLock.Unlock()
}

// RecordReadRequest records that a read is requested for a key,
// no read record is created.
func RecordReadRequest(key string) {
	keyReadsLock.Lock()
	reads, ok := keyReads[key]
	if !ok {
		reads = 0
	}
	keyReads[key] = reads + 1
	keyReadsLock.Unlock()
}

// RecordRead records a read for key which returned value at NOW
func RecordReadNow(key string, value []byte) {
	RecordRead(key, value, time.Now().UnixNano())
}

// CheckWriteForAnomaly checks a write of value to key at a time when
func CheckWriteForAnomaly(key string, value []byte, when int64) bool {
	keyReadRecordsLock.Lock()
	readRecord, found := keyReadRecords[key]
	// Whether or not it's an anomaly, we don't want to check again (no double counting)
	delete(keyReadRecords, key)
	keyReadRecordsLock.Unlock()
	if !found {
		return false
	}

	// If the write is in the past and it's different than
	// what was returned on the read then it's an anomaly
	//
	// currently using a 'grace period' of 35 ms
	// as in the existential consistency paper
	// think if we should use something else
	// (like measure clock skew ourselves?
	if readRecord.whenReadNano-35000000 > when {
		if !bytes.Equal(readRecord.valueRead, value) {
			registerAnomaly(key)
			return true
		}
	}

	return false
}

// CheckWriteForAnomaly checks a write of value to key at NOW
func CheckWriteForAnomalyNow(key string, value []byte) bool {
	return CheckWriteForAnomaly(key, value, time.Now().UnixNano())
}

// registerAnomaly registers an anomaly for key
func registerAnomaly(key string) {
	keyAnomaliesLock.Lock()
	val, found := keyAnomalies[key]
	if !found {
		val = 0
	}
	val++
	keyAnomalies[key] = val
	keyAnomaliesLock.Unlock()
}

// AnomalyCountForKey returns the anomaly count for key
func AnomalyCountForKey(key string) int {
	keyAnomaliesLock.Lock()
	count, ok := keyAnomalies[key]
	keyAnomaliesLock.Unlock()
	if !ok {
		return 0
	}
	return count
}

// ReadCountForKey returns the total number of reads performed for key
func ReadCountForKey(key string) int {
	keyReadsLock.Lock()
	count, ok := keyReads[key]
	if !ok {
		return 0
	}
	keyReadsLock.Unlock()
	return count
}

// AnomalyRateForKey returns the anomaly percentage, as an integer >= 0 and <= 100,
// for a specific key.
//
// The rate is calculated over the total number of reads.
func AnomalyRateForKey(key string) int {
	keyReadsLock.Lock()
	totalReads, ok := keyReads[key]
	keyReadsLock.Unlock()
	if !ok || totalReads == 0 {
		return 0
	}
	keyAnomaliesLock.Lock()
	anomalyCount, ok := keyAnomalies[key]
	keyAnomaliesLock.Unlock()
	if !ok {
		return 0
	}
	return int(float64(anomalyCount) * 100.0 / float64(totalReads))
}
