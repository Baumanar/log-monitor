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

func main() {
	// Create a global context used by the monitor and the display for cancellation signals
	ctx, cancel := context.WithCancel(context.Background())

	// Flags of the app
	isDemo := flag.Bool("flagname", true, "demo or not")
	logFile := flag.String("logfile", "/tmp/access.log", "demo or not")
	timeWindow := flag.Int("time window", 120, "time window for alerting")
	threshold := flag.Int("threshold", 120, "threshold for alerting")
	updateFreq := flag.Int("update frequency", 5, "update frequency of the statistics")

	if _, err := os.Stat(*logFile); os.IsNotExist(err) {
		log.Fatal(fmt.Sprintf("file %s does not exist.", *logFile))
	}

	// Channel to display statistics
	displayChan := make(chan monitoring.StatRecord)
	// Channel to alert
	alertChan := make(chan monitoring.AlertRecord)

	// Create a new monitor and a new display with the given parameters
	monitor := monitoring.New(ctx, *logFile, displayChan, alertChan, *timeWindow, *updateFreq, *threshold)
	display := display.New(ctx, cancel, displayChan, alertChan)

	// If the app is running in demo mode, write concurrently logs to the log file
	if *isDemo {
		// Get a random seed
		rand.Seed(time.Now().UnixNano())
		// Write logs in a goroutine
		go monitoring.LogGenerator(ctx, *logFile)
	}

	// Run the monitor in a goroutine
	go monitor.Run()

	// Do the displaying
	display.Run()

}
