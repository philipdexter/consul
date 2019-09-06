package reflective

import (
	"bytes"
	"fmt"
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
var keyAnomalies map[string]*movingAverage = map[string]*movingAverage{}
var keyAnomaliesLock = sync.Mutex{}

// keyReads stores the total number of reads
// performed for each key
var keyReads map[string]*movingAverage = map[string]*movingAverage{}
var keyReadsLock = sync.Mutex{}

const (
	movingWindow = iota
	decayingAverage
)

var reflectiveAdapting = movingWindow

func InitConfig(refladapting string) error {
	if refladapting == "moving_window" {
		reflectiveAdapting = movingWindow
	} else if refladapting == "decaying_average" {
		reflectiveAdapting = decayingAverage
	} else {
		return fmt.Errorf("invalid reflective adapting")
	}

	return nil
}

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
	ma, ok := keyReads[key]
	if !ok {
		ma = newMovingAverage()
		keyReads[key] = ma
	}
	ma.increment()
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
	ma, found := keyAnomalies[key]
	if !found {
		ma = newMovingAverage()
		keyAnomalies[key] = ma
	}
	ma.increment()
	keyAnomaliesLock.Unlock()
}

// AnomalyCountForKey returns the anomaly count for key
func AnomalyCountForKey(key string) int {
	keyAnomaliesLock.Lock()
	defer keyAnomaliesLock.Unlock()
	ma, ok := keyAnomalies[key]
	if !ok {
		return 0
	}
	return ma.total()
}

// ReadCountForKey returns the total number of reads performed for key
func ReadCountForKey(key string) int {
	keyReadsLock.Lock()
	defer keyReadsLock.Unlock()
	ma, ok := keyReads[key]
	if !ok {
		return 0
	}
	return ma.total()
}

// AnomalyRateForKey returns the anomaly percentage, as an integer >= 0 and <= 100,
// for a specific key.
//
// The rate is calculated over the total number of reads.
func AnomalyRateForKey(key string) int {
	totalReads := 0
	keyReadsLock.Lock()
	ma, ok := keyReads[key]
	if ok {
		totalReads = ma.total()
	}
	keyReadsLock.Unlock()
	if totalReads == 0 {
		return 0
	}

	anomalyCount := 0
	keyAnomaliesLock.Lock()
	ma, ok = keyAnomalies[key]
	if ok {
		anomalyCount = ma.total()
	}
	keyAnomaliesLock.Unlock()
	if anomalyCount == 0 {
		return 0
	}
	return int(float64(anomalyCount) * 100.0 / float64(totalReads))
}

///

func getNow() int64 {
	return time.Now().Unix()
}

type movingAverage struct {
	ma        [5]int
	last_time int64
}

func newMovingAverage() *movingAverage {
	return &movingAverage{}
}

func (ma *movingAverage) increment() {
	now := getNow()
	second := int(now % 5)

	diff := int(now - ma.last_time)
	if diff > 5 {
		diff = 5
	}
	for i := 0; i < diff; i++ {
		// fmt.Println("overwriting", (second+i)%5)
		ma.ma[(second+i)%5] = 0
	}

	ma.ma[second] += 1
	// fmt.Println("current second is", second)
	// fmt.Println("values are", ma)
	// fmt.Println("value second is", ma.ma[second], "total is", ma.total())

	ma.last_time = now
}

func (ma *movingAverage) total() int {
	agg := 0
	for _, val := range ma.ma {
		agg += val
	}
	return agg
}
