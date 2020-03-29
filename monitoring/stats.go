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


func getStats(records []LogRecord, k int) StatRecord {
	sectionMap := make(map[string]int, 0)
	methodMap := make(map[string]int, 0)
	statusMap := make(map[string]int, 0)
	requests := len(records)
	bytesCount := 0

	for _, log := range records{
		sectionMap[log.section]++
		methodMap[log.method]++
		statusMap[log.status]++
		bytesCount += log.bytes
	}
	return StatRecord{
		TopSections: getTopK(sectionMap, k),
		TopMethods:  getTopK(methodMap, k),
		TopStatus:   getTopK(statusMap, k),
		NumRequests: requests,
		BytesCount:  bytesCount,
	}
}

func Min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

type Pair struct {
	Key   string
	Value int
}

func getTopK(countMap map[string]int, k int) []Pair {
	var ss []Pair
	for k, v := range countMap {
		ss = append(ss, Pair{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	return ss[:Min(k, len(ss))]
}
