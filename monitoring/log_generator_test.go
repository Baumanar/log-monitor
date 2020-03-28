package monitoring

import (
	"fmt"
	"testing"
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
