package monitoring

import (
	"context"
	"fmt"
	"github.com/Baumanar/log-monitor/pkg/generator"
	"log"
	"os"
	"reflect"
	"testing"
	"time"
)

// TestLogMonitor_readLog tests the ReadLog function
func TestLogMonitor_readLog(t *testing.T) {
	// Create or empty the test log file

	tests := []struct {
		name string
		want int
	}{
		//{"test0", 1},
		{"test1", 10},
		{"test2", 456},
		{"test3", 789},
		{"test4", 499},
		{"test5", 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			_, err := os.Create("test.log")
			if err != nil {
				log.Fatal(err)
			}

			//Write some lines in the log file
			for i := 0; i < 50; i++ {
				generator.WriteLogLine("test.log")
			}

			// Create a new monitor
			statChan := make(chan StatRecord)
			alertChan := make(chan AlertRecord)
			monitor := New(ctx, cancel, "test.log", statChan, alertChan, 10, 5, 10)

			go func() {
				// Let a short time for the monitor to get at the end of the file
				time.Sleep(100 * time.Millisecond)
				// Write the log file
				for i := 0; i < tt.want; i++ {
					generator.WriteLogLine("test.log")
				}
				// Sleep for a short time to let the monitor compute and finish
				time.Sleep(200 * time.Millisecond)
				// Stop the monitor
				cancel()
			}()

			// Check for new lines
			monitor.ReadLog()

			if len(monitor.LogRecords) != tt.want {
				t.Errorf("ReadLog() \nread = %v lines \nwant %v lines", len(monitor.LogRecords), tt.want)
			}
		})
	}
	err := os.Remove("test.log")
	if err != nil {
		log.Fatal(err)
	}

}

// Checks if the ReadLog is able to exit when the cancellation function is called
func TestLogMonitor_readLog1(t *testing.T) {

	t.Run("test cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		_, err := os.Create("test.log")
		if err != nil {
			log.Fatal(err)
		}
		// Create a new monitor
		statChan := make(chan StatRecord)
		alertChan := make(chan AlertRecord)
		monitor := New(ctx, cancel, "test.log", statChan, alertChan, 10, 5, 10)
		// Wait for 1 second before cancelling
		go func() {
			time.Sleep(time.Second * 1)
			cancel()
		}()
		// run the reading
		monitor.ReadLog()
	})

	err := os.Remove("test.log")
	if err != nil {
		log.Fatal(err)
	}

}

func TestLogMonitor_alert(t *testing.T) {

	tests := []struct {
		name          string
		threshold     int
		timeWindow    int
		alertTraffics [][]int
		// Alert state of the monitor at the beginning
		startState bool
		want       []AlertRecord
	}{
		// Going Above the threshold and staying above
		{"test0",
			10,
			120,
			[][]int{{10, 10, 10}, {500, 700, 500}, {1000, 500, 600}},
			false,
			[]AlertRecord{{Alert: true, NumTraffic: 1700}}},

		// Going Below the threshold and staying below
		{"test1",
			10,
			120,
			[][]int{{500, 500, 500}, {10, 10, 10}, {100, 100, 100}},
			true, []AlertRecord{{Alert: false, NumTraffic: 30}}},

		// Always below the threshold
		{"test2",
			10,
			120,
			[][]int{{10, 10, 10}, {40, 30, 20}, {50, 60, 70}},
			false,
			[]AlertRecord{}}, // This value will not be compared to the output ouf the channel as it won't send anything

		// Always above
		{"test3", 10,
			120,
			[][]int{{500, 500, 500}, {600, 600, 600}, {600, 600, 600}},
			true,
			[]AlertRecord{}}, // This value will not be compared to the output ouf the channel as it won't send anything

		// First below, then above for a few steps, then below
		{"test4", 10, 120,
			[][]int{{200, 0, 500}, {1000, 1000, 1000}, {600, 800, 800}, {600, 600, 600}, {600, 100, 1000}, {100, 100, 100}},
			false,
			[]AlertRecord{{Alert: true, NumTraffic: 3000}, {Alert: false, NumTraffic: 300}}}, // should send two alerts
	}

	// Run tests
	ctx, cancel := context.WithCancel(context.Background())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new monitor
			statChan := make(chan StatRecord)
			alertChan := make(chan AlertRecord)
			monitor := New(ctx, cancel, "test.log", statChan, alertChan, tt.timeWindow, 5, tt.threshold)
			// Init the Alert state
			monitor.InAlert = tt.startState
			go func() {
				for _, traffic := range tt.alertTraffics {
					monitor.AlertTraffic = traffic
					monitor.Alert()
				}
				close(alertChan)
			}()

			// Index of the current wanted Alert
			idx := 0
			// Listen to the Alert channel
			for {
				select {
				case got, ok := <-alertChan:
					if ok {
						if got != tt.want[idx] {
							t.Errorf("ReadLog() \ngot = %v lines \nwant %v lines", got, tt.want)
						}
						// go to the next wanted Alert
						idx++
					} else { // If the channel is closed, return
						return
					}
				}
			}
		})
	}
}

func TestLogMonitor_report(t *testing.T) {

	tests := []struct {
		name       string
		logRecords []LogRecord
		want       StatRecord
	}{
		{"test0",
			[]LogRecord{
				{"a", "a", "a", "a", "a", "a", "a", "a", 2000},
				{"a", "a", "a", "a", "a", "a", "a", "a", 2000},
				{"a", "a", "a", "a", "a", "a", "a", "a", 5000},
				{"b", "b", "b", "b", "b", "b", "b", "b", 10000}},

			StatRecord{[]Pair{{"a", 3}, {"b", 1}},
				[]Pair{{"a", 3}, {"b", 1}},
				[]Pair{{"a", 3}, {"b", 1}},
				4,
				"19.0 kB"},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			statChan := make(chan StatRecord)
			alertChan := make(chan AlertRecord)
			monitor := New(ctx, cancel, "test.log", statChan, alertChan, 120, 5, 10)
			go func() {
				monitor.LogRecords = tt.logRecords
				monitor.Report()
			}()
			got := <-monitor.StatChan
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Report() \ngot = %v \nwant %v", got, tt.want)
			}
			if monitor.LogRecords != nil {
				t.Errorf("LogRecords is not empty")

			}
		})
	}
}

// Checks if the monitor is able to exit when the cancellation function is called
func TestLogMonitor_Run(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"cancel_test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := os.Create("test.log")
			if err != nil {
				log.Fatal(err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			// Make buffered channels of size 3 so they are not blocking, we won't use them here
			statChan := make(chan StatRecord, 3)
			alertChan := make(chan AlertRecord, 3)
			// Set the alertFreq to 1 second so the function still sends some info the the statChan
			monitor := New(ctx, cancel, "test.log", statChan, alertChan, 120, 1, 10)

			// call cancel after 2 seconds
			go func() {
				time.Sleep(2 * time.Second)
				cancel()
			}()
			// Start log generation, if should be stopped after 1s
			monitor.Run()
		})
	}
	err := os.Remove("test.log")
	if err != nil {
		log.Fatal(err)
	}
}

// Checks if alertTraffic reaches maximum size and does not goes over this size
func TestLogMonitor_Run1(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"cancel_test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := os.Create("test.log")
			if err != nil {
				log.Fatal(err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			// Make buffered channels of size 3 so they are not blocking, we won't use them here
			statChan := make(chan StatRecord)
			alertChan := make(chan AlertRecord)
			// The size of the alertTraffic should be maximum 3 and be updated every second
			monitor := New(ctx, cancel, "test.log", statChan, alertChan, 3, 1, 10)
			monitor.LogRecords = []LogRecord{}
			go func() {
				// Let the monitor run for 5 seconds
				ticker := time.NewTicker(time.Second * time.Duration(6))
				// counter ton compare the size of AlertTraffic a each tick
				counter := 1
				for {
					select {
					case <-monitor.StatChan:
						//fmt.Println(counter % 3, monitor.AlertIndex)
						if counter%3 != monitor.AlertIndex {
							t.Errorf("monitor.LogRecords should be maximum %d: got: %d", Min(counter, 3), len(monitor.AlertTraffic))
						}
						counter++

					case <-ticker.C:
						cancel()
					}
				}
			}()
			// Start log generation, if should be stopped after 1s
			monitor.Run()
		})
	}
	err := os.Remove("test.log")
	if err != nil {
		log.Fatal(err)
	}
}

// Checks if the monitor is able to exit when the cancellation function is called
func TestLogMonitor_Run2(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"cancel_test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := os.Create("test.log")
			if err != nil {
				log.Fatal(err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			// Make buffered channels of size 3 so they are not blocking, we won't use them here
			statChan := make(chan StatRecord)
			alertChan := make(chan AlertRecord)
			// Set the alertFreq to 1 second so the function still sends some info the the statChan
			monitor := New(ctx, cancel, "test.log", statChan, alertChan, 120, 1, 1000000)

			count := 0
			go func() {
				//stop writing logs after 10 seconds
				tickerWrite := time.NewTicker(time.Duration(10) * time.Second)

				time.Sleep(500 * time.Millisecond)
				step := 2000
				for {
					select {
					case <-tickerWrite.C:
						return
					default:
						for i := 0; i < step; i++ {
							generator.WriteLogLine("test.log")
						}
						count += step
						time.Sleep(50 * time.Millisecond)
						//fmt.Printf("generator wrote %d lines\n", count)
					}
				}
			}()
			// Start log generation, if should be stopped after 1s
			countRead := 0
			go func() {
				//stop reading logs after 12 seconds
				tickerRead := time.NewTicker(time.Duration(12) * time.Second)
				for {

					select {
					case stat := <-statChan:
						// add the traffic number to the AlertTraffic array
						countRead += stat.NumRequests
						fmt.Printf("monitor read %d lines\n", countRead)
					case alert := <-alertChan:
						// add the traffic number to the AlertTraffic array
						fmt.Printf("Alert:  %d \n", alert.NumTraffic)
					case <-tickerRead.C:
						// Cancel to stop the monitor running in the main process
						cancel()
					}
				}
			}()
			monitor.Run()
			fmt.Printf("Final written: %d \n Final read: %d \n", count, countRead)
			if countRead != count {
				t.Errorf("Final written: %d \n Final read: %d \n", count, countRead)
			}
		})
	}
	err := os.Remove("test.log")
	if err != nil {
		log.Fatal(err)
	}
}
