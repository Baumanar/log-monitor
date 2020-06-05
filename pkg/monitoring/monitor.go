package monitoring

import (
	"context"
	"github.com/hpcloud/tail"
	"log"
	"sync"
	"time"
)

const sleepTime = 50

// LogMonitor listens to the log file and retrieves new logs
// Computes the statistics of the new logs and sends them to the display
// Sends Alert whenever the threshold is exceeded or recovers
type LogMonitor struct {
	// The log file to read
	LogFile string
	// Time window for the alerting in seconds
	// TimeWindow*Threshold gives the maximum number of logs before alerting (by default 120*10=1200 logs in 2min)
	TimeWindow int
	// number of seconds between each send to the display
	UpdateInterval int
	// Current Alert status
	InAlert bool
	// Maximum number of request per second before alerting
	Threshold int
	// Current LogRecords
	LogRecords []LogRecord
	// Number of requests at each update, used for alerting
	AlertTraffic []int
	AlertIndex   int
	// mutex for thread safety
	Mutex sync.Mutex
	// channel to communicate statistics to the display
	StatChan chan StatRecord
	// channel to send alerts to the display
	AlertChan chan AlertRecord
	// Reopen the file if truncated
	ReOpenFile bool
	// Global app context
	ctx    context.Context
	cancel context.CancelFunc
}

// New returns a new LogMonitor with the specified parameters
func New(ctx context.Context, cancel context.CancelFunc, logFile string, statChan chan StatRecord, alertChan chan AlertRecord, timeWindow int, updateInterval int, threshold int, ReOpenFile bool) *LogMonitor {
	var mutex sync.Mutex
	monitor := &LogMonitor{
		LogFile:        logFile,
		TimeWindow:     timeWindow,
		UpdateInterval: updateInterval,
		InAlert:        false,
		Threshold:      threshold,
		LogRecords:     make([]LogRecord, 0),
		AlertTraffic:   make([]int, timeWindow/updateInterval),
		AlertIndex:     0,
		Mutex:          mutex,
		StatChan:       statChan,
		AlertChan:      alertChan,
		ctx:            ctx,
		cancel:         cancel,
		ReOpenFile:     ReOpenFile,
	}
	return monitor
}

// ReadLog reads reads the log file
// continuously  checks for new log lines
func (m *LogMonitor) ReadLog() {
	// To continuously read the log file, we use the package tail (github.com/hpcloud/tail) that mimicks the fail -f behavior
	// this package also manages file truncation/rotation which is nice
	tailListener, err := tail.TailFile(m.LogFile, tail.Config{Follow: true, ReOpen: m.ReOpenFile, MustExist: true, Logger: tail.DiscardingLogger})
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case <-m.ctx.Done():
			return
		case line := <-tailListener.Lines:

			newRecord, err := ParseLogLine(line.Text)
			// Thread safety, add new logRecords
			// Lock to avoid that the monitor flushes the array at the same time when sending statistics
			// If the log has been correctly parsed, add it to the current record list
			if err == nil {
				m.Mutex.Lock()
				m.LogRecords = append(m.LogRecords, *newRecord)
				m.Mutex.Unlock()
			} else {
				log.Fatal(err)
			}
		}
	}
}

// Alert sends alerts to the display by sending an AlertRecord to the display through the Alert channel
func (m *LogMonitor) Alert() {
	numTraffic := 0
	// sum up the traffic in the time window
	for _, t := range m.AlertTraffic {
		numTraffic += t
	}
	// If the number of requests is above threshold*timeWindow and the monitor was not in Alert
	// set InAlert to true and send an AlertRecord to the display
	if numTraffic > m.Threshold*m.TimeWindow && !m.InAlert {
		m.InAlert = true
		m.AlertChan <- AlertRecord{
			Alert:      true,
			NumTraffic: numTraffic,
		}
		// If the number of requests is below threshold*timeWindow and the monitor was in in Alert
		// set InAlert to false and send an AlertRecord to the display
	} else if numTraffic < m.Threshold*m.TimeWindow && m.InAlert {
		m.InAlert = false
		m.AlertChan <- AlertRecord{
			Alert:      false,
			NumTraffic: numTraffic,
		}
	}
}

// Report sends log statistics to the display
func (m *LogMonitor) Report() {
	// Compute the stats of the current records
	m.Mutex.Lock()
	statRecord := GetStats(m.LogRecords, 5)

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
	go m.ReadLog()
	// Do the alerting and send the statistics each UpdateInterval seconds with a ticker
	ticker := time.NewTicker(time.Second * time.Duration(m.UpdateInterval))
	for {
		select {
		case <-ticker.C:
			// add the traffic number to the AlertTraffic array
			m.AlertTraffic[m.AlertIndex] = len(m.LogRecords)
			m.AlertIndex += 1
			m.AlertIndex = m.AlertIndex % (m.TimeWindow / m.UpdateInterval)
			m.Alert()
			m.Report()
		case <-m.ctx.Done():
			return
		}
	}
}
