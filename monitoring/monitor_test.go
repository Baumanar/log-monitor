package monitoring

import (
	"context"
	"log"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"
)

// TestLogMonitor_readLog tests the readLog function
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
				writeLogLine("test.log")
			}

			// Create a new monitor
			displayChan := make(chan StatRecord)
			alertChan := make(chan AlertRecord)
			monitor := New(ctx, cancel,"test.log", displayChan, alertChan, 10, 5, 10)

			// Check for new lines

			go monitor.readLog()
			// Let a short time for the monitor to get at the end of the file
			time.Sleep(500 * time.Millisecond)

			// Write the log file
			for i := 0; i < tt.want; i++ {
				writeLogLine("test.log")
			}
			// Sleep for a short time to let the monitor compute and finish
			time.Sleep(1000 * time.Millisecond)

			if len(monitor.LogRecords) != tt.want {
				t.Errorf("readLog() \nread = %v lines \nwant %v lines", len(monitor.LogRecords), tt.want)
			}

			err = os.Remove("test.log")
			if err != nil{
				log.Fatal(err)
			}
		})
	}

}

func TestLogMonitor_readLog1(t *testing.T) {

	t.Run("test cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())


		_, err := os.Create("test.log")
		if err != nil {
			log.Fatal(err)
		}

		// Create a new monitor
		displayChan := make(chan StatRecord)
		alertChan := make(chan AlertRecord)
		monitor := New(ctx, cancel,"test.log", displayChan, alertChan, 10, 5, 10)

		// run the reading
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			go monitor.readLog()
			wg.Done()
		}()
		time.Sleep(time.Millisecond*50)
		go func() {
			cancel()
			wg.Done()
		}()
		wg.Wait()
	})
}


func TestLogMonitor_alert(t *testing.T) {

	tests := []struct {
		name          string
		threshold     int
		timeWindow    int
		alertTraffics [][]int
		// alert state of the monitor at the beginning
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
			displayChan := make(chan StatRecord)
			alertChan := make(chan AlertRecord)
			monitor := New(ctx, cancel, "test.log", displayChan, alertChan, tt.timeWindow, 5, tt.threshold)
			// Init the alert state
			monitor.InAlert = tt.startState
			go func() {
				for _, traffic := range tt.alertTraffics {
					monitor.AlertTraffic = traffic
					monitor.alert()
				}
				close(alertChan)
			}()

			// Index of the current wanted alert
			idx := 0
			// Listen to the alert channel
			for {
				select {
				case got, ok := <-alertChan:
					if ok {
						if got != tt.want[idx] {
							t.Errorf("readLog() \ngot = %v lines \nwant %v lines", got, tt.want)
						}
						// go to the next wanted alert
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
				{"a", "a", "a", "a", "a", "a", "a", "a", 2},
				{"a", "a", "a", "a", "a", "a", "a", "a", 2},
				{"a", "a", "a", "a", "a", "a", "a", "a", 2},
				{"b", "b", "b", "b", "b", "b", "b", "b", 2}},

			StatRecord{[]Pair{{"a", 3}, {"b", 1}},
				[]Pair{{"a", 3}, {"b", 1}},
				[]Pair{{"a", 3}, {"b", 1}},
				4,
				8},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			displayChan := make(chan StatRecord)
			alertChan := make(chan AlertRecord)
			monitor := New(ctx, cancel, "test.log", displayChan, alertChan, 120, 5, 10)
			go func() {
				monitor.LogRecords = tt.logRecords
				monitor.report()
			}()
			got := <-monitor.StatChan
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("report() \ngot = %v \nwant %v", got, tt.want)
			}
			if monitor.LogRecords != nil {
				t.Errorf("LogRecords is not empty")

			}
		})
	}
}

func TestLogMonitor_Run(t *testing.T) {
	type fields struct {
		LogFile      string
		TimeWindow   int
		UpdateFreq   int
		InAlert      bool
		Threshold    int
		LogRecords   []LogRecord
		AlertTraffic []int
		Mutex        sync.Mutex
		StatChan     chan StatRecord
		AlertChan    chan AlertRecord
		ctx          context.Context
		cancel       context.CancelFunc
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LogMonitor{
				LogFile:      tt.fields.LogFile,
				TimeWindow:   tt.fields.TimeWindow,
				UpdateFreq:   tt.fields.UpdateFreq,
				InAlert:      tt.fields.InAlert,
				Threshold:    tt.fields.Threshold,
				LogRecords:   tt.fields.LogRecords,
				AlertTraffic: tt.fields.AlertTraffic,
				Mutex:        tt.fields.Mutex,
				StatChan:     tt.fields.StatChan,
				AlertChan:    tt.fields.AlertChan,
				ctx:          tt.fields.ctx,
				cancel:       tt.fields.cancel,
			}
		})
	}
}