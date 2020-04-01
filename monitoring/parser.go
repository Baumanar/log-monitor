package monitoring

import (
	"errors"
	"regexp"
	"strconv"
)

// LogRecord gathers all the information about a parsed log line
type LogRecord struct {
	// Remote hostname or IP number
	remotehost string
	// The remote logname of the user
	rfc931 string
	// The username as which the user has authenticated himself.
	authuser string
	// Date and time of the request
	date string
	// Http verb used
	method string
	// Section of the request
	section string
	// The HTTP status code returned to the client
	status string
	// The HTTP protocol version
	protocol string
	// The content-length of the document transferred
	bytesCount int
}

// Compile the regex once and use it for every log line
var regex = regexp.MustCompile(`(\S+)\s+(\S+)\s+(\S+)\s+(\[.+\])\s+\"([A-Z]+)\s+(\/[^\/]+)\/.+\s+(\S+)\"\s+(\S+)\s+([0-9]+|-)(.+)?`)

// Parses a log record according to the w3c-formatted HTTP access log and return the LogRecord associated
func ParseLogLine(input string) (*LogRecord, error) {
	// log pattern

	matches := regex.FindStringSubmatch(input)
	// if the log record is badly formatted, return an empty record as well as an error
	if len(matches) != 11 {
		return nil, errors.New("Invalid log format.")
	}
	var bytes int
	if matches[9] != "-"{
		bytes, _ = strconv.Atoi(matches[9])
	} else {
		bytes = 0
	}


	// return a new LogRecord instance
	return &LogRecord{
		remotehost: matches[1],
		rfc931:     matches[2],
		authuser:   matches[3],
		date:       matches[4],
		method:     matches[5],
		section:    matches[6],
		protocol:   matches[7],
		status:     matches[8],
		bytesCount: bytes,
	}, nil
}
