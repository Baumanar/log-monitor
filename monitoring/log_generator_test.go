package monitoring

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"math/rand"
	"os"
	"testing"
)



func Test_generateLog(t *testing.T) {
	rand.Seed(10)

	tests := []struct {
		name string
		want string
	}{
		{"test0", "110.240.59.99 - john [31/March/2020:11:56:45 +0200] \"DELETE /cart/books?id=321 HTTP/1.0\" 403 618\n"},
		{"test1", "193.5.206.103 - john [31/March/2020:11:56:45 +0200] \"DELETE /about/user?id=22&checkout=True HTTP/1.0\" 202 7284\n"},
		{"test2", "110.56.80.71 - paul [31/March/2020:11:56:45 +0200] \"PATCH /about/view.html HTTP/1.0\" 401 1114\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateLog(); got != tt.want {
				t.Errorf("generateLog() = \n%v%v", got, tt.want)
			}
		})
	}
}
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