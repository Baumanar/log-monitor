package monitoring

import (
	"fmt"
	"sort"
)

func getStats(records []LogRecord) (map[string]int, map[string]int, map[string]int, int, int) {
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
	return sectionMap, methodMap,statusMap,requests, bytesCount
}

func Min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

type kv struct {
	Key   string
	Value int
}

func getTopK(countMap map[string]int, k int) string{
	var ss []kv
	for k, v := range countMap {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})
	var out string
	for i:=0; i<Min(k, len(ss));i++ {
		out += fmt.Sprintf("%s, %d\n", ss[i].Key, ss[i].Value)
	}
	return out
}
