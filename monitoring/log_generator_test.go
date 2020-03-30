package monitoring

import (
	"fmt"
	"testing"
	"time"
)



func Test_generateLog(t *testing.T) {
	for i:=0;i<10;i++{
		got := generateLog()
		fmt.Println(got)
		if got == "" {
			t.Errorf("generateLog() = %v", got)
		}
	}

}

func TestRandSinSleep(t *testing.T) {
	type args struct {
		step     float64
		variance float64
	}

		// TODO: Add test cases.

	for i:=0;i<300;i++ {
		a := RandSinSleep(time.Now().Unix(), 0)
		fmt.Println(i, a, int(a), 0)
		//fmt.Println(time.Now().Unix())
		time.Sleep(time.Second*time.Duration(int(a)))
	}
}
