package monitoring

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

type LogMonitor struct{
	LogFile      string
	TimeWindow     int
	UpdateFreq	   int
	InAlert        bool
	Threshold      int
	InvalidRecords int
	LogRecords     []LogRecord
	AlertTraffic   []int
	Mutex          sync.Mutex
	OutputChan 	   chan StatRecord
	AlertChan	   chan string
	Ctx 		   context.Context
}

func (m *LogMonitor) Init(logFile string, displayChan chan StatRecord, alertChan chan string, timeWindow int, ctx context.Context){
	m.LogFile = logFile
	m.TimeWindow = timeWindow
	m.UpdateFreq = 5
	m.InAlert = false
	m.Threshold = 10
	m.InvalidRecords = 0
	m.LogRecords = make([]LogRecord, 0)
	m.Mutex = sync.Mutex{}
	m.OutputChan = displayChan
	m.AlertChan = alertChan
	m.Ctx = ctx
}

func NewMonitor(logFile string, displayChan chan StatRecord, alertChan chan string) (LogMonitor){
	var mutex sync.Mutex
	monitor := LogMonitor{
		LogFile: logFile,
		TimeWindow:   3,
		UpdateFreq: 5,
		InAlert:      false,
		Threshold:    10,
		InvalidRecords: 0,
		LogRecords:     make([]LogRecord, 0),
		AlertTraffic: make([]int, 0),
		Mutex: mutex,
		OutputChan: displayChan,
		AlertChan: alertChan,
	}
	return monitor
}

func (m *LogMonitor) readLog() {
	file, _ := os.Open(m.LogFile)
	defer file.Close()
	_, err := file.Seek(0, 2)
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot find %s file.", m.LogFile))
	}
	reader := bufio.NewReader(file)
	var line string
	for {
		select {
		case <-m.Ctx.Done():
			return
		default:
			line, err = reader.ReadString('\n')
			if err == io.EOF || line =="" {
				time.Sleep(time.Millisecond*50)
			} else if err != io.EOF {
				newRecord, err := parseLogLine(line)
				m.Mutex.Lock()
				if err != nil{
					m.InvalidRecords++
				} else {
					m.LogRecords = append(m.LogRecords, newRecord)
				}
				m.Mutex.Unlock()
			} else {
				log.Fatal(err)
			}
		}
	}
}

func (m *LogMonitor) addTraffic(traffic int){
	m.AlertTraffic = append(m.AlertTraffic, traffic)
}

func (m *LogMonitor) removeTraffic(){
	m.AlertTraffic = m.AlertTraffic[1:]
}

func (m *LogMonitor) alert(){
	numTraffic := 0
	for _,t := range m.AlertTraffic {
		numTraffic += t
	}
	if numTraffic > m.Threshold*m.TimeWindow && !m.InAlert {
		m.InAlert = true
		m.AlertChan <- fmt.Sprintf("\n High traffic generated an alert - hits = %d, triggered at %s\n", numTraffic, time.Now().Format("15:04:05, January 02 2006"))
	} else if  numTraffic < m.Threshold*m.TimeWindow && m.InAlert {
		m.InAlert = false
		m.AlertChan <- fmt.Sprintf("High traffic has recovered, triggered at %s\n", time.Now().Format("15:04:05, January 02 2006"))
	}
}

func (m *LogMonitor) report(){
	out := make([]string, 0)
	out = append(out, fmt.Sprintf("Number of records: %d Invalid:  %d\n", len(m.LogRecords), m.InvalidRecords))
	statRecord := getStats(m.LogRecords, 5)

	m.Mutex.Lock()
	m.LogRecords = nil
	m.InvalidRecords = 0
	m.Mutex.Unlock()
	m.OutputChan <- statRecord
}



func (m *LogMonitor) Run() {
	// Concurrently red the log file
	go m.readLog()
	// Do the alerting and send the statistics
	for{
		time.Sleep(time.Duration(m.UpdateFreq)*time.Second)
		m.addTraffic(len(m.LogRecords))
		if len(m.AlertTraffic)>m.TimeWindow {
			m.removeTraffic()
		}
		m.alert()
		m.report()
	}
}

