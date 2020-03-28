package monitoring

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"
)



// Lists of string useful for fake log generation
var users = []string{"james", "jill", "frank", "mary", "john", "paul", "jennifer", "sarah"}
var sections = []string{"/home", "/products", "/about", "/api", "/contact", "/profile", "/report", "/posts", "/login", "/cart" }
var subsections = []string{"/view.html", "/request/865/", "/ref=lh_cart", "/books?id=321", "/user?id=123", "/register", "/user?id=22&checkout=True"}
var verbs = []string{"POST", "GET", "PUT", "PATCH", "DELETE"}
var status = []string{"200", "201", "202", "203", "204", "300", "301", "302", "400", "401", "402", "403", "404", "500", "501", "502", "503"}

// Generates a random IP adress
func randomIP() string{
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
}
// Generates a random HTTP request
func randomRequest() string{
	return fmt.Sprintf("\"%s %s%s HTTP/1.0\"", verbs[rand.Intn(len(verbs))], sections[rand.Intn(len(sections))], subsections[rand.Intn(len(subsections))])
}
// Generates a random user among the fixed user list
func randomUser() string{
	return fmt.Sprintf("%s", users[rand.Intn(len(users))])
}

// Generates a random user among the fixed user list
func randomStatus() string{
	return fmt.Sprintf("%s", status[rand.Intn(len(status))])
}

// Generates a random byte size
func randomByteSize() string{
	return fmt.Sprintf("%d", rand.Intn(10000))
}
// Returns current time formatted in the proper format
func currentTime() string {
	return time.Now().Format("[02/January/2006:15:04:05 -0700]")
}

// Generates the full log line
func generateLog() string{
	return fmt.Sprintf("%s - %s %s %s %s %s\n",  randomIP(), randomUser(), currentTime(), randomRequest(), randomStatus(), randomByteSize())
}

// Writes a generated log line in the log file
func WriteLogLine(logFile string){
	// Open file in append mode to write log lines at the end of the file
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err = f.WriteString(generateLog()); err != nil {
		panic(err)
	}
}


func LogGenerator(logFile string, ctx context.Context){
	var min, max int64
	min, max= 50, 300

	r := min + rand.Int63n(max)
	ticker := time.NewTicker(time.Duration(r)*time.Millisecond)
	for {
		select {
			case <-ticker.C:
				//fmt.Println("ticker ticked, wirtting new log")
				WriteLogLine(logFile)
				r := min + rand.Int63n(max)
				ticker = time.NewTicker(time.Duration(r)*time.Millisecond)

			case <-ctx.Done():
				return
			}
		}
	}
