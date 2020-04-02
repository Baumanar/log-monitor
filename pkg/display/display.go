package display

import (
	"context"
	"fmt"
	"github.com/Baumanar/log-monitor/pkg/monitoring"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/sparkline"
	"github.com/mum4k/termdash/widgets/text"
	"log"
	"time"
)

// Display displays information to the terminal
type Display struct {
	// StatChan is the channel receiving statistic information from the monitor
	StatChan chan monitoring.StatRecord
	// AlertChan is the channel receiving statistic information from the monitor
	AlertChan chan monitoring.AlertRecord
	// termdash text displaying the uptime
	uptimeDisplay *text.Text
	// termdash text displaying the statistics
	statDisplay *text.Text
	// alert text displaying the statistics
	alertDisplay *text.Text
	// histogram of the number of requests received
	// the histogram does not show the number of requests, it just shows the evolution of the traffic
	histogram *sparkline.SparkLine
	// Global app context
	ctx context.Context
	// cancel function for context cancellation
	cancel context.CancelFunc
}

// New returns a new Display with the specified parameters
func New(ctx context.Context, cancel context.CancelFunc, statChan chan monitoring.StatRecord, alertChan chan monitoring.AlertRecord) *Display {
	// Initialize displays
	uptimeDisplay, err := text.New(text.WrapAtWords())
	if err != nil {
		log.Fatal(err)
	}
	statDisplay, err := text.New(text.WrapAtWords())
	if err != nil {
		log.Fatal(err)
	}
	alertDisplay, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		log.Fatal(err)
	}
	histogram, err := sparkline.New(sparkline.Color(cell.ColorCyan))
	if err != nil {
		log.Fatal(err)
	}

	display := &Display{
		StatChan:      statChan,
		AlertChan:     alertChan,
		uptimeDisplay: uptimeDisplay,
		statDisplay:   statDisplay,
		alertDisplay:  alertDisplay,
		histogram:     histogram,
		ctx:           ctx,
		cancel:        cancel,
	}
	return display
}

// DisplayPairs displays statistic pairs to the statDisplay
func (d *Display) DisplayPairs(pairs []monitoring.Pair) {
	// We iterate i and not on the elements of pairs to always have the same number of lines printed
	for i := 0; i < 5; i++ {
		// display each pair on a row
		if i < len(pairs) {
			d.statDisplay.Write(fmt.Sprintf("    %s: %d\n", pairs[i].Key, pairs[i].Value))
		} else {
			// display an empty line
			d.statDisplay.Write(fmt.Sprintf("\n"))
		}
	}
}

// DisplayInfo displays all the information on the statDisplay:
// 		The pairs of each section/method/status
// 		The number of requests
// 		The number of bytes
func (d *Display) DisplayInfo(stat monitoring.StatRecord) {

	d.statDisplay.Write(fmt.Sprintf("Number of requests: "), text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
	d.statDisplay.Write(fmt.Sprintf("%d\n", stat.NumRequests))
	d.statDisplay.Write(fmt.Sprintf("Number of bytes transferred: "), text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
	d.statDisplay.Write(fmt.Sprintf("%s\n", stat.BytesCount))

	d.statDisplay.Write("\nTop sections: \n", text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
	d.DisplayPairs(stat.TopSections)

	d.statDisplay.Write("Top sections: \n", text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
	d.DisplayPairs(stat.TopMethods)

	d.statDisplay.Write("Top status: \n", text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
	d.DisplayPairs(stat.TopStatus)
}

// FmtDuration formats the uptime
func FmtDuration(d time.Duration) string {
	// Convert duration to int
	// Quite inconvenient, but I did not find a way to keep it to duration
	uptime := int(d) / 1000000000
	s := uptime % 60
	m := uptime / 60 % 60
	h := uptime / 3600
	return fmt.Sprintf("%02dh%02dmin%02ds", h, m, s)
}

// Update updates all panels at once
func (d *Display) Update(ctx context.Context) {
	startTime := time.Now().Round(time.Second)
	ticker := time.NewTicker(time.Second)
	for {
		select {
		// Update uptime each second
		case <-ticker.C:
			d.uptimeDisplay.Reset()
			d.uptimeDisplay.Write(fmt.Sprintf("%s", FmtDuration(time.Since(startTime).Round(time.Second))))
			// New statistics received
		case info, ok := <-d.StatChan:
			if ok {
				// add the number of requests to the histogram
				d.histogram.Add([]int{info.NumRequests})
				// Clear the past information
				d.statDisplay.Reset()
				// Display new information
				d.DisplayInfo(info)
			} else {
				d.cancel()
			}
			// Alert received
		case alert, ok := <-d.AlertChan:
			if ok {
				if alert.Alert {
					// If alert is true, display it in red
					d.alertDisplay.Write(fmt.Sprintf("High traffic generated an alert - hits = %d, triggered at %s\n", alert.NumTraffic, time.Now().Format("15:04:05, January 02 2006")), text.WriteCellOpts(cell.FgColor(cell.ColorRed)))

				} else {
					// If the alert recovered, display it in green
					d.alertDisplay.Write(fmt.Sprintf("High traffic has recovered, triggered at %s\n", time.Now().Format("15:04:05, January 02 2006")), text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
				}
			} else {
				d.cancel()
			}
		case <-ctx.Done():
			return
		}
	}
}

// Run is the main function of the Display
func (d *Display) Run() {
	// Run a goroutine to update the all panels
	go d.Update(d.ctx)

	// If q is pressed, exit
	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			d.cancel()
		}
	}
	// Create a new global box
	box, err := termbox.New()
	if err != nil {
		log.Fatal(err)
	}
	// Create containers
	// Containers are responsible for the layout of the dashboard
	container, err := container.New(
		box,
		container.Border(linestyle.Light),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.SplitVertical(
			container.Left(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle("Uptime"),
						container.PlaceWidget(d.uptimeDisplay)),
					container.Bottom(
						container.Border(linestyle.Light),
						container.BorderTitle("Traffic info"),
						container.PlaceWidget(d.statDisplay)),
					container.SplitPercent(15),
				)),
			container.Right(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle("Alerts"),
						container.PlaceWidget(d.alertDisplay)),
					container.Bottom(
						container.Border(linestyle.Light),
						container.BorderTitle("Traffic histogram"),
						container.PlaceWidget(d.histogram)),
				),
			),
		))
	// Defer the closing
	defer box.Close()

	if err != nil {
		log.Fatal(err)
	}

	// Run the dashboard
	termdash.Run(d.ctx, box, container, termdash.KeyboardSubscriber(quitter))
}
