package display

import (
	"context"
	"fmt"
	"github.com/Baumanar/log-monitor/monitoring"
	"strconv"
	"testing"
	"time"
)

func Test_fmtDuration(t *testing.T) {
	type args struct {
		d time.Duration
	}
	tests := []struct {
		name  string
		input time.Duration
		want  string
	}{
		{"test0", time.Duration(time.Second * 3610), "01h00min10s"},
		{"test1", time.Duration(time.Second * 4652), "01h17min32s"},
		{"test2", time.Duration(time.Second * 79846), "22h10min46s"},
		{"test3", time.Duration(time.Second * 365), "00h06min05s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmtDuration(tt.input); got != tt.want {
				t.Errorf("fmtDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}


// TestDisplay_Run tests the overall functionality of the display
// We send fake stats and alerts through the channels and check that the displayer displays the correct information
// Unfortunately the test is only visual as I did not find a way to make more robust tests
func TestDisplay_Run(t *testing.T) {
	tests := []struct {
		name   string
	}{
		{"test0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			statChan := make(chan monitoring.StatRecord)
			alertChan := make(chan monitoring.AlertRecord)
			display := New(ctx, cancel, statChan, alertChan)

			// Send 30 times stats and alerts
			go func() {
				for i:=0;i<30;i++{
					statChan <- monitoring.StatRecord{
						TopSections: []monitoring.Pair{{"/ten", i%10},
							{"/five", i%5},
							{"/three", i%3}},
						TopMethods:  []monitoring.Pair{{"TEN", i%10},
							{"FIVE", i%5},
							{"THREE", i%3}},
						TopStatus:   []monitoring.Pair{{"1xx", i%10},
							{"5xx", i%5},
							{"3xx", i%3}},
						NumRequests: i,
						BytesCount:  strconv.Itoa(i)+".0 kB",
					}
					alertChan <- monitoring.AlertRecord{
						Alert:      i%2==0,
						NumTraffic: i,
					}
					time.Sleep(200*time.Millisecond)
				}
				close(alertChan)
				close(statChan)
			}()
			// Run the display, It should stop when the channels are closed
			display.Run()

			fmt.Println("Ok")




		})
	}
}