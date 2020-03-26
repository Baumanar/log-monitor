package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"
)


type logMonitor struct{
	logs chan []string
	tenSecs int
	inAlert bool
	threshold int
	traffic [][]string
}


func (m *logMonitor) addTraffic(traffic[]string){
	m.traffic = append(m.traffic, traffic)
}

func (m *logMonitor) pop(){
	m.traffic = m.traffic[1:]
}

func (m *logMonitor) alert(){
	numTraffic := 0
	for _,t := range m.traffic{
		numTraffic += len(t)
	}
	if numTraffic > m.threshold*m.tenSecs && !m.inAlert{
		m.inAlert = true
		fmt.Printf("High traffic generated an alert - hits = %d, triggered at %s\n", numTraffic, time.Now().Format("15:04:05, January 02 2006"))
	} else if  numTraffic < m.threshold*m.tenSecs && m.inAlert {
		m.inAlert = false
		fmt.Printf("High traffic has recovered, triggered at %s\n", time.Now().Format("15:04:05, January 02 2006"))
	}
}

func (m *logMonitor) run() {
	for {
		select{
		case tr, ok := <-m.logs:
			if ok{
				m.addTraffic(tr)
				if len(m.traffic)>m.tenSecs{
					m.pop()
				}
				m.alert()
			} else {
				return
			}
		default:
		}
	}
}

func writeLogLine(logFile string, text string){
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err = f.WriteString(RandStringRunes(5)+"\n"); err != nil {
		panic(err)
	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}



func readLog(logFile string, done chan bool, out chan string) {
	file, _ := os.Open(logFile)
	defer file.Close()
	_, err := file.Seek(0, 2)
	if err != nil {
		log.Fatal("aaa", err)
	}

	reader := bufio.NewReader(file)
	var line string
	cont := true
	for cont{
		select {
		case <-done:
			cont = false
			close(out)

		default:
			line, err = reader.ReadString('\n')
			if err != io.EOF {
				//fmt.Printf(" > Read %d characters\n", len(line))
				out <- line
				//fmt.Println(" > > " + line)
			}
		}
	}
}



func gatherLogs(inLog chan string, outLog chan []string){
	ticker := time.NewTicker(time.Second * time.Duration(2))
	logList := make([]string, 0)
	cont := true
	for cont{
		select {
		case <-ticker.C:
			fmt.Println("Ticker ticked")
			fmt.Println("number of logs found: ", len(logList))
			outLog <- logList
			logList = nil

		case log, ok := <-inLog:
			logList = append(logList, log)
			if !ok {
				cont = false
			}
		}
	}


}




func main(){
	rand.Seed(time.Now().UnixNano())
	isDemo := flag.Bool("flagname", true, "demo or not")
	var min, max int64
	min, max= 0, 4500
	done := make(chan bool)
	inLogs := make(chan string)
	outLogs := make(chan []string)


	monitor := logMonitor{
		logs:        outLogs,
		inAlert:     false,
		threshold:   700,
		tenSecs: 3,
		traffic:     make([][]string, 0),
	}

	if *isDemo{
		go func() {
			for {
				r := min + rand.Int63n(max)
				//fmt.Println(r)
				time.Sleep(time.Duration(r)*time.Microsecond)
				writeLogLine("test.log", "aaa")
			}
			done <- true
		}()
	}


	go gatherLogs(inLogs, outLogs)
	go monitor.run()
	readLog("test.log", done, inLogs)
	close(done)
}






