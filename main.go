package main

import (
	"context"
	"flag"
	"log-monitor/display"
	"log-monitor/monitoring"
	"math/rand"
	_ "net/http/pprof"
	"time"
)






func main(){

	//ctx, cancel := context.WithCancel(context.Background())
	rand.Seed(time.Now().UnixNano())
	isDemo := flag.Bool("flagname", true, "demo or not")
	logFile := flag.String("logfile", "/tmp/access.log", "demo or not")

	ctx, cancel := context.WithCancel(context.Background())


	displayChan := make(chan string)
	alertChan := make(chan string)

	monitor := monitoring.LogMonitor{}
	monitor.Init(*logFile, displayChan, alertChan, 3)




	if *isDemo{
		go monitoring.LogGenerator(*logFile, ctx)
	}

	//go monitoring.GatherLogs(inLogs, outLogs)
	go monitor.Run()


	displayer := display.Displayer{}
	displayer.Init(displayChan, alertChan, ctx, cancel)
	displayer.Run()
}






