package monitoring

import (
	"fmt"
	"sort"
)

// StatRecord is the type passed from the Monitor to
// the display when there is an update
type StatRecord struct {
	TopSections []Pair
	TopMethods  []Pair
	TopStatus   []Pair
	NumRequests int
	// the number of byte send will already be formatted
	BytesCount string
}

// AlertRecord is the type passed from the Monitor to
// the display when there is an alert
// if Alert is true, the threshold has been exceeded
// if Alert in false, the alert recovered
// NumTraffic is the current number of request in the timeWindow (2min default)
type AlertRecord struct {
	Alert      bool
	NumTraffic int
}

// Pair is composed by a Key and a Value
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
	var bytesCount int

	// update maps and byteCount with each record
	for _, log := range records {
		sectionMap[log.section]++
		methodMap[log.method]++
		statusMap[processStatus(log.status)]++
		bytesCount += log.bytesCount
	}
	return StatRecord{
		TopSections: getTopK(sectionMap, k),
		TopMethods:  getTopK(methodMap, k),
		TopStatus:   getTopK(statusMap, k),
		NumRequests: requests,
		BytesCount:  formatByteCount(bytesCount),
	}
}

// Min returns the min between to integers
func Min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

// processStatus converts a status to its first number and converts the rest to x
func processStatus(status string) string {
	if len(status) == 3 {
		return status[:1] + "xx"
	}
	return status
}

// Format byte count into kB/MB/GB
func formatByteCount(b int) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := unit, 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

// getTopK return k pairs of (Key, Value) whose values are the highest in the countMap map
// Returns an ordered array of Pair of size K, the first element of the array has the highest value
// This method could be improved in terms of performance, but as we have a limited number of sections/methods/status
// it remains effective
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
