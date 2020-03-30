package monitoring

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type LogRecord struct{
	remotehost string
	rfc931 string
	authuser string
	date string
	method string
	section string
	status string
	protocol string
	bytes int

}

// Parses a log record according to the w3c-formatted HTTP access log
func parseLogLine(input string) (LogRecord, error){
	// log pattern
	r := regexp.MustCompile(`(?P<remotehost>\S+) (?P<rfc931>\S+) (?P<authuser>\S+) (?P<date>\[.+\]) \"(?P<request>[A-Z]+) (\/\S+)\/.+ (\S+)\" (?P<status>\S+) (?P<bytes>\S+)(.+)?`)
	matches := r.FindStringSubmatch(input)
	// if the log record is badly formatted, return an empty record as well as an error
	if len(matches) != 11 {
		return LogRecord{} , errors.New("Invalid log format.")
	} else{
		// return a new LogRecord instance
		bytes, err := strconv.Atoi(matches[9])
		if err != nil {
			fmt.Println(err)
		}
		return LogRecord{
			remotehost: matches[1],
			rfc931:     matches[2],
			authuser:   matches[3],
			date:       matches[4],
			method:    	matches[5],
			section:	matches[6],
			protocol:   matches[7],
			status:     matches[8],
			bytes:      bytes,
		}, nil
	}
}

