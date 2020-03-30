package monitoring

import (
	"sort"
)

type StatRecord struct{
	TopSections []Pair
	TopMethods []Pair
	TopStatus[]Pair
	NumRequests int
	BytesCount int
}

// An AlertRecord is the type passed from the Monitor to
// the display when there is an alert
// if Alert is true, the threshold has been exceeded
// if Alert in false, the alert recovered
// NumTraffic is the current number of request in the timeWindow (2min default)
type AlertRecord struct{
	Alert bool
	NumTraffic int
}

// A pair is composed by a Key and a Value
// Key is the name of the section/method/status
// Value is the number of hits
type Pair struct {
	Key   string
	Value int
}

// Computes the statistics from a list of LogRecords records
// k is number of lines for each stat to display
// Returns a statRecord with the top sections/HTTP methods/status, the number of requests and the number of bytes
func getStats(records []LogRecord, k int) StatRecord {

	// Create maps to count the number of hits for sections, HTTP methods and status
	// Get the number of requests during this step (10s default) and initialize a byte count
	sectionMap := make(map[string]int, 0)
	methodMap := make(map[string]int, 0)
	statusMap := make(map[string]int, 0)
	requests := len(records)
	bytesCount := 0

	// update maps and byteCount with each record
	for _, log := range records{
		sectionMap[log.section]++
		methodMap[log.method]++
		statusMap[log.status]++
		bytesCount += log.bytesCount
	}
	return StatRecord{
		TopSections: getTopK(sectionMap, k),
		TopMethods:  getTopK(methodMap, k),
		TopStatus:   getTopK(statusMap, k),
		NumRequests: requests,
		BytesCount:  bytesCount,
	}
}

// Min returns the min between to integers
func Min(x, y int) int {
	if x > y {
		return y
	}
	return x
}


// getTopK return k pairs of (Key, Value) whose values are the highest in the countMap map
// Returns an ordered array of Pair of size K, the first element of the array has the highest value
// This
func getTopK(countMap map[string]int, k int) []Pair {
	// Create an array of Pair with the countMap
	var pairs []Pair
	for k, v := range countMap {
		pairs = append(pairs, Pair{k, v})
	}
	// Sort the array
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Value > pairs[j].Value
	})
	// Return a slice of the sorted array
	// if k > len(pairs), return the whole array
	return pairs[:Min(k, len(pairs))]
}
