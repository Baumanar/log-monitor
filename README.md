# HTTP Log-monitor

A simple console program that monitors HTTP traffic on your machine. Listens to an actively written log file and displays statistics

## Requirements

The project is written in Go and uses go modules. Developpement was made on go1.14.1 but was also tested on go1.13.

### Build

To build the project, simply run:
```sh
  go build
```

### Run

Once you built the project run:

```sh
  ./log-monitor
```

The options of the program are the following:


```
Usage of ./log-monitor:
  -demo
    	demo or not, if demo the log file will be concurrently written with fake logs
  -logfile string
    	logfile path (default "/tmp/access.log")
  -threshold int
    	threshold for alerting in requests per second (default 10)
  -timewindow int
    	time window for alerting in seconds (default 120)
  -updatefreq int
    	number of seconds between each statistic update (default 10)
```
type ./log-monitor -help to display this message.

Example:
```sh
./log-monitor -logfile /tmp/access.log -threshold 10 -timewindow 60 -updatefreq 5
```

## Demo
If the ```demo``` flag is set to true, a separate goroutine writes the log file to simulate logging.

## Architecture

The architecture of the log-monitor has two main components:
- A monitor
- A display

The monitor communicates with the display by using two channels: one for statistics, one for alerts

The monitor listens to the log file and continuously checks for new logs. It keeps trace of the logs 
written during the last ```updatefreq``` seconds. Every ```updatefreq``` it computes statistics of the current 
logs and sends the computed statistics to the display by using the statistics channel. The statistics sent are:
- The 5 most requested sections
- The 5 most used  HTTP methods
- The 5 most frequent HTTP status codes returned
- The number of requests
- The number of bytes transferred

The monitor also checks for alerts, 
if the average traffic during the last ```timewindow``` exceeds the threshold per second, an alert is sent to the display. 
Alerts are sent by using the alert channel

The displayer contains 4 pannels:
- The uptime of the program
- The information pannel on which statistics are displayed
- The alert pannel on which alerts are displayed
- An histogram of the traffic evolution. The purpose of the histogram is just to give an intuition of the traffic evolution

## Improvements

The app could be improved in many ways:
- The monitor continuously checks for file modification, it would be better that file modifications trigger events. This could be maybe done by some package like 
[fsnotify](https://github.com/fsnotify/fsnotify) or [tail](https://github.com/hpcloud/tail)
- Improve the user interface. If I had more time, I think I would have implemented an interface available on a web browser for conviency






