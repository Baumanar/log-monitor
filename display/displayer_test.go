package display

import (
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