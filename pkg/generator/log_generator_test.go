package generator

import (
	"bufio"
	"bytes"
	"context"
	"github.com/hpcloud/tail"
	"io"
	"log"
	"os"
	"testing"
	"time"
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

// Test_writeLogLine checks if the number of lines written by the generator is correct
func Test_writeLogLine(t *testing.T) {

	tests := []struct {
		name    string
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
			for i := 0; i < tt.lineNum; i++ {
				WriteLogLine("test.log")
			}
			file, _ := os.Open("test.log")
			reader := bufio.NewReader(file)
			// Count the number of lines written
			count, err := lineCounter(reader)
			if err != nil {
				t.Errorf("ReadLog() \nwrote = %v lines \nwant %v lines", count, tt.lineNum)
			}
		})
	}

	err := os.Remove("test.log")
	if err != nil {
		log.Fatal(err)
	}
}

// Test_writeLogLine checks if the number of lines written by the generator is correct
func Test_writeLogLine1(t *testing.T) {

	tests := []struct {
		name    string
		lineNum int
	}{
		{"test0", 10},
		{"test0", 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create or empty file
			_, err := os.Create("test1.log")
			if err != nil {
				log.Fatal(err)
			}
			// Write a certain number of lines
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				for i := 0; i < tt.lineNum; i++ {
					WriteLogLine("test1.log")
				}
				time.Sleep(2 * time.Second)
				cancel()
			}()

			count := 0
			tail_text, err := tail.TailFile("test1.log", tail.Config{Follow: true, ReOpen: true, MustExist: true})
			if err != nil {
				log.Fatal(err)
			}

			for {
				select {
				case <-tail_text.Lines:
					count++
				case <-ctx.Done():
					return
				}
			}

			if err != nil {
				t.Errorf("ReadLog() \nwrote = %v lines \nwant %v lines", count, tt.lineNum)
			}
		})
	}

	err := os.Remove("test1.log")
	if err != nil {
		log.Fatal(err)
	}
}

// Checks if the generator is able to exit when the cancellation function is called
func TestLogGenerator(t *testing.T) {
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

			// call cancel after 1s
			go func() {
				time.Sleep(1 * time.Second)
				cancel()
			}()
			// Start log generation, if should be stopped after 1s
			LogGenerator(ctx, "test.log", 10)
		})
	}
	err := os.Remove("test.log")
	if err != nil {
		log.Fatal(err)
	}
}
