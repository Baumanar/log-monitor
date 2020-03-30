package monitoring

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

const sleepTime = 50


// LogMonitor listens to the log file and retrieves new logs
// Computes the statistics of the new logs and sends them to the display
// Sends alert whenever the threshold is exceeded or recovers
type LogMonitor struct{
	// The log file to read
	LogFile      string
	// Time window for the alreting
	TimeWindow     int
	// number of seconds between each send to the display
	UpdateFreq	   int
	// Current alert status
	InAlert        bool
	// Maximum number of request per second before alerting
	Threshold      int
	InvalidRecords int
	// Current LogRecords
	LogRecords     []LogRecord
	// Number of requests at each update, used for alerting
	AlertTraffic   []int
	// mutex for thread safety
	Mutex          sync.Mutex
	// channel to communicate statistics to the display
	StatChan 	   chan StatRecord
	// channel to send alerts to the display
	AlertChan	   chan AlertRecord
	// Global app context
	Ctx 		   context.Context
}

// Returns a new LogMonitor with the specified parameters
func New(logFile string, displayChan chan StatRecord, alertChan chan AlertRecord, timeWindow int, updateFreq int, threshold int, ctx context.Context) *LogMonitor {
	var mutex sync.Mutex
	monitor := &LogMonitor{
		LogFile: logFile,
		TimeWindow:   timeWindow,
		UpdateFreq: updateFreq,
		InAlert:      false,
		Threshold:    threshold,
		InvalidRecords: 0,
		LogRecords:     make([]LogRecord, 0),
		AlertTraffic: make([]int, 0),
		Mutex: mutex,
		StatChan: displayChan,
		AlertChan: alertChan,
		Ctx: ctx,
	}
	return monitor
}

// readLog reads reads the log file
// continuously  checks for new log lines
func (m *LogMonitor) readLog() {
	// Open the file and defer closing when returning
	file, _ := os.Open(m.LogFile)
	defer file.Close()
	// Go to the end of the file to get newest log lines
	_, err := file.Seek(0, 2)
	if err != nil {
		log.Fatal()
		m.Ctx.Done()
	}

	// Create a buffer for reading
	reader := bufio.NewReader(file)
	var line string
	for {
		select {
		case <-m.Ctx.Done():
			return
		default:
			// Read the file to the end of the current line
			line, err = reader.ReadString('\n')
			// If no new line was found, sleep for a short time to avoid overcomputing
			if err == io.EOF || line =="" {
				time.Sleep(time.Millisecond*sleepTime)
			} else if err != io.EOF {
				// A new line has been found,  parse if to create a new logRecord
				newRecord, err := parseLogLine(line)
				// Thread safety, add new logRecords
				// Lock to avoid that the monitor flushes the array at the same time when sending statistics
				m.Mutex.Lock()
				if err != nil{
					m.InvalidRecords++
				} else {
					m.LogRecords = append(m.LogRecords, *newRecord)
				}
				m.Mutex.Unlock()
			} else {
				log.Fatal(err)
			}
		}
	}
}


// alert sends alerts to the display by sending an AlertRecord to the display through the Alert channel
func (m *LogMonitor) alert(){
	numTraffic := 0
	// sum up the traffic in the time window
	for _,t := range m.AlertTraffic {
		numTraffic += t
	}
	// If the number of requests exceeds the threshold and the monitor was not in alert
	// set InAlert to true and send an AlertRecord to the display
	if numTraffic > m.Threshold*m.TimeWindow && !m.InAlert {
		m.InAlert = true
		m.AlertChan <- AlertRecord{
			Alert:      true,
			NumTraffic: numTraffic,
		}
	// If the number of requests is below the threshold and the monitor was in in alert
	// set InAlert to false and send an AlertRecord to the display
	} else if  numTraffic < m.Threshold*m.TimeWindow && m.InAlert {
		m.InAlert = false
		m.AlertChan <- 	AlertRecord{
			Alert:      false,
			NumTraffic: numTraffic,
		}
	}
}

//  Report sends log statistics to the display
func (m *LogMonitor) report(){
	// Compute the stats of the current records
	statRecord := getStats(m.LogRecords, 5)
	//
	m.Mutex.Lock()
	// Thread safety, add new logRecords
	// Lock to avoid that the monitor adds new records at the same time it is flushing
	m.LogRecords = nil
	m.InvalidRecords = 0
	m.Mutex.Unlock()
	// Send stats using the StatCha
	m.StatChan <- statRecord
}



func (m *LogMonitor) Run() {
	// Concurrently red the log file
	go m.readLog()
	// Do the alerting and send the statistics each UpdateFreq seconds
	for{
		time.Sleep(time.Duration(m.UpdateFreq)*time.Second)
		// add the traffic number to the AlertTraffic array
		m.AlertTraffic = append(m.AlertTraffic, len(m.LogRecords))
		// If the length of the array is bigger than the window, remove the oldest traffic number
		if len(m.AlertTraffic)>m.TimeWindow {
			m.AlertTraffic = m.AlertTraffic[1:]
		}
		m.alert()
		m.report()
	}
}

