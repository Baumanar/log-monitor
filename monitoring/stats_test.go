package monitoring

import "testing"

func Test_processStatus(t *testing.T) {
	type args struct {
		status string
	}
	tests := []struct {
		name string
		input string
		want string
	}{
		{"test0", "400", "4xx"},
		{"test1", "200", "2xx"},
		{"test2", "3xx", "3xx"},
		{"test3", "", ""},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := processStatus(tt.input); got != tt.want {
				t.Errorf("processStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMin(t *testing.T) {
	type args struct {
		x int
		y int
	}
	tests := []struct {
		name string
		inputX int
		inputY int
		want int
	}{
		{"test0", 3,2,2},
		{"test0", 2,3,2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Min(tt.inputX, tt.inputY); got != tt.want {
				t.Errorf("Min() = %v, want %v", got, tt.want)
			}
		})
	}
}