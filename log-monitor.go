package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Baumanar/log-monitor/display"
	"github.com/Baumanar/log-monitor/monitoring"
	"log"
	"math/rand"
	"os"
	"time"
)

const startInterval = 4000.0

func main() {
	// Create a global context used by the monitor and the display for cancellation signals
	ctx, cancel := context.WithCancel(context.Background())

	// Flags of the app
	isDemo := flag.Bool("demo", false, "demo or not, if demo the log file will be concurrently written with fake logs")
	logFile := flag.String("logfile", "/tmp/access.log", "logfile path")
	timeWindow := flag.Int("timewindow", 120, "time window for alerting in seconds")
	threshold := flag.Int("threshold", 10, "threshold for alerting in requests per second")
	updateFreq := flag.Int("updatefreq", 10, "number of seconds between each statistic update")
	flag.Parse()

	// Verify that the log file exists
	if _, err := os.Stat(*logFile); os.IsNotExist(err) {
		log.Fatal(fmt.Sprintf("file %s does not exist.", *logFile))
	}

	// Channel to display statistics
	statChan := make(chan monitoring.StatRecord)
	// Channel to alert
	alertChan := make(chan monitoring.AlertRecord)

	// Create a new monitor and a new display with the given parameters
	monitor := monitoring.New(ctx, cancel,*logFile, statChan, alertChan, *timeWindow, *updateFreq, *threshold)
	display := display.New(ctx, cancel, statChan, alertChan)

	// If the app is running in demo mode, write concurrently logs to the log file
	if *isDemo {
		// Get a random seed
		rand.Seed(time.Now().UnixNano())
		// Write logs in a goroutine
		go monitoring.LogGenerator(ctx, *logFile, startInterval)
	}

	// Run the monitor in a goroutine
	go monitor.Run()

	// Do the displaying
	display.Run()

}

