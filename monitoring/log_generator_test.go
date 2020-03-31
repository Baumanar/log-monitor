package monitoring

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"testing"
)


// Helper Func for Test_writeLogLine
func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}


func Test_writeLogLine(t *testing.T) {
	type args struct {
		logFile string
	}
	tests := []struct {
		name string
		lineNum int
	}{
		{"test0", 10},
		{"test0", 100},
		{"test0", 1000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create or empty file
			_, err := os.Create("test.log")
			if err != nil {
				log.Fatal(err)
			}
			// Write a certain number of lines
			for i:=0;i<tt.lineNum;i++{
				writeLogLine("test.log")
			}
			file, _ := os.Open("test.log")
			reader := bufio.NewReader(file)
			// Count the number of lines written
			count, err := lineCounter(reader)
			if err != nil{
				t.Errorf("readLog() \nwrote = %v lines \nwant %v lines", count, tt.lineNum)
			}
		})
	}
}