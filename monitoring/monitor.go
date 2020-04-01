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
type LogMonitor struct {
	// The log file to read
	LogFile string
	// Time window for the alerting in seconds
	// TimeWindow*Threshold gives the maximum number of logs before alerting (by default 120*10=1200 logs in 2min)
	TimeWindow int
	// number of seconds between each send to the display
	UpdateFreq int
	// Current alert status
	InAlert bool
	// Maximum number of request per second before alerting
	Threshold int
	// Current LogRecords
	LogRecords []LogRecord
	// Number of requests at each update, used for alerting
	AlertTraffic []int
	// mutex for thread safety
	Mutex sync.Mutex
	// channel to communicate statistics to the display
	StatChan chan StatRecord
	// channel to send alerts to the display
	AlertChan chan AlertRecord
	// Global app context
	ctx    context.Context
	cancel context.CancelFunc
}

// New returns a new LogMonitor with the specified parameters
func New(ctx context.Context, cancel context.CancelFunc, logFile string, statChan chan StatRecord, alertChan chan AlertRecord, timeWindow int, updateFreq int, threshold int) *LogMonitor {
	var mutex sync.Mutex
	monitor := &LogMonitor{
		LogFile:      logFile,
		TimeWindow:   timeWindow,
		UpdateFreq:   updateFreq,
		InAlert:      false,
		Threshold:    threshold,
		LogRecords:   make([]LogRecord, 0),
		AlertTraffic: make([]int, 0),
		Mutex:        mutex,
		StatChan:     statChan,
		AlertChan:    alertChan,
		ctx:          ctx,
		cancel:       cancel,
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
	}

	// Create a buffer for reading
	reader := bufio.NewReader(file)
	var line string
	for {
		select {
		case <-m.ctx.Done():
			return
		default:
			// Read the file to the end of the current line
			line, err = reader.ReadString('\n')
			// If no new line was found, sleep for a short time to avoid over computing
			if err == io.EOF {
				time.Sleep(time.Millisecond * sleepTime)
			} else if err != io.EOF {
				// A new line has been found, parse if to create a new logRecord
				newRecord, err := parseLogLine(line)
				// Thread safety, add new logRecords
				// Lock to avoid that the monitor flushes the array at the same time when sending statistics
				m.Mutex.Lock()
				// If the log has been correctly parsed, add it to the current record list
				if err == nil {
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
func (m *LogMonitor) alert() {
	numTraffic := 0
	// sum up the traffic in the time window
	for _, t := range m.AlertTraffic {
		numTraffic += t
	}
	// If the number of requests is above threshold*timeWindow and the monitor was not in alert
	// set InAlert to true and send an AlertRecord to the display
	if numTraffic > m.Threshold*m.TimeWindow && !m.InAlert {
		m.InAlert = true
		m.InAlert = true
		m.AlertChan <- AlertRecord{
			Alert:      true,
			NumTraffic: numTraffic,
		}
		// If the number of requests is below threshold*timeWindow and the monitor was in in alert
		// set InAlert to false and send an AlertRecord to the display
	} else if numTraffic < m.Threshold*m.TimeWindow && m.InAlert {
		m.InAlert = false
		m.AlertChan <- AlertRecord{
			Alert:      false,
			NumTraffic: numTraffic,
		}
	}
}

//  Report sends log statistics to the display
func (m *LogMonitor) report() {
	// Compute the stats of the current records
	statRecord := getStats(m.LogRecords, 5)
	//
	m.Mutex.Lock()
	// Thread safety, add new logRecords
	// Lock to avoid that the monitor adds new records at the same time it is flushing
	m.LogRecords = nil
	m.Mutex.Unlock()
	// Send stats using the StatChan
	m.StatChan <- statRecord
}

// Run is the main function of the monitor
func (m *LogMonitor) Run() {
	// Concurrently read the log file
	go m.readLog()
	// Do the alerting and send the statistics each UpdateFreq seconds with a ticker
	ticker := time.NewTicker(time.Second * time.Duration(m.UpdateFreq))
	for {
		select {
		case <-ticker.C:
			// add the traffic number to the AlertTraffic array
			m.AlertTraffic = append(m.AlertTraffic, len(m.LogRecords))
			// If the length of the array is bigger than the window/updateFreq, remove the oldest traffic number
			if len(m.AlertTraffic) > (m.TimeWindow / m.UpdateFreq) {
				m.AlertTraffic = m.AlertTraffic[1:]
			}
			m.alert()
			m.report()
		case <-m.ctx.Done():
			return
		}
	}
}
