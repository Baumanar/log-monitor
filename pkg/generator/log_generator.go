package generator

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

// Lists of string useful for fake log generation
var users = []string{"james", "jill", "frank", "mary", "john", "paul", "jennifer", "sarah"}
var sections = []string{"/home", "/products", "/about", "/api", "/contact", "/profile", "/Report", "/posts", "/login", "/cart"}
var subsections = []string{"/view.html", "/request/865/", "/ref=lh_cart", "/books?id=321", "/user?id=123", "/register", "/user?id=22&checkout=True"}
var verbs = []string{"POST", "GET", "PUT", "PATCH", "DELETE"}
var status = []string{"200", "201", "202", "203", "204", "300", "301", "302", "400", "401", "402", "403", "404", "500", "501", "502", "503"}

// RandomIP generates a random IP address
func RandomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
}

// RandomRequest generates a random HTTP request
func RandomRequest() string {
	return fmt.Sprintf("\"%s %s%s HTTP/1.0\"", verbs[rand.Intn(len(verbs))], sections[rand.Intn(len(sections))], subsections[rand.Intn(len(subsections))])
}

// RandomUser generates a random user among the fixed user list
func RandomUser() string {
	return fmt.Sprintf("%s", users[rand.Intn(len(users))])
}

// RandomStatus generates a random user among the fixed user list
func RandomStatus() string {
	return fmt.Sprintf("%s", status[rand.Intn(len(status))])
}

// RandomByteSize generates a random byte size
func RandomByteSize() string {
	return fmt.Sprintf("%d", rand.Intn(10000))
}

// CurrentTime returns current time formatted in the proper format
func CurrentTime() string {
	return time.Now().Format("[02/January/2006:15:04:05 -0700]")
}

// GenerateLog generates the full log line
func GenerateLog() string {
	return fmt.Sprintf("%s - %s %s %s %s %s\n", RandomIP(), RandomUser(), CurrentTime(), RandomRequest(), RandomStatus(), RandomByteSize())
}

// WriteLogLine writes a generated log line in the log file
func WriteLogLine(logFile string) {
	// Open file in append mode to write log lines at the end of the file
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if _, err = f.WriteString(GenerateLog()); err != nil {
		log.Fatal(err)
	}
}

// LogGenerator generates logs and writes them to the log file
// The evolution of the generation rate follows as triangle in order to generate alerts
func LogGenerator(ctx context.Context, logFile string, startInterval float64) {

	// addVal will increment the counter
	addVal := 1.0
	// Count is the number with which the ticker interval varies
	// range: 10 to 100
	count := 5.0

	// the ticker duration changes at each tick
	ticker := time.NewTicker(time.Duration(startInterval) * time.Millisecond)

	for {
		select {
		case <-ticker.C:
			// When tick, write a log line
			WriteLogLine(logFile)
			rand := rand.Float64() * 50
			ticker.Stop()
			// change the ticker duration
			ticker = time.NewTicker(time.Duration(startInterval/count+rand) * time.Millisecond)
			if count > 100 {
				addVal = -0.1
			}
			if count < 10 {
				addVal = +0.1
			}
			count += addVal
		case <-ctx.Done():
			return
		}
	}
}
