package display

import (
	"context"
	"fmt"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/sparkline"
	"github.com/mum4k/termdash/widgets/text"
	"log-monitor/monitoring"
	"time"
)

type Displayer struct {
	StatChan chan monitoring.StatRecord
	AlertChan chan monitoring.AlertRecord
	uptimeDisplay *text.Text
	statDisplay *text.Text
	alertDisplay *text.Text
	histogram *sparkline.SparkLine
	ctx context.Context
	cancel context.CancelFunc
}




func New(statChan chan monitoring.StatRecord, alertChan chan monitoring.AlertRecord, ctx context.Context, cancel context.CancelFunc) *Displayer{
	uptimeDisplay, err := text.New(text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	statDisplay, err := text.New(text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	alertDisplay, err := text.New(text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	histogram, err := sparkline.New(sparkline.Color(cell.ColorCyan))
	if err != nil {
		panic(err)
	}


	displayer := &Displayer{
		StatChan:      statChan,
		AlertChan:     alertChan,
		uptimeDisplay: uptimeDisplay,
		statDisplay:   statDisplay,
		alertDisplay:  alertDisplay,
		histogram:     histogram,
		ctx:           ctx,
		cancel:        cancel,
	}

	return displayer
}


func (d *Displayer) displayPairs( pairs []monitoring.Pair) {
	for i:=0;i<5;i++{
		if i < len(pairs){
			err := d.statDisplay.Write(fmt.Sprintf("    %s: %d\n", pairs[i].Key, pairs[i].Value))
			if err != nil {
				panic(err)
			}
		} else {
			err := d.statDisplay.Write(fmt.Sprintf("\n"))
			if err != nil {
				panic(err)
			}
		}
	}
}

func (d *Displayer) displayInfo (stat monitoring.StatRecord){
	if err := d.statDisplay.Write("\nTop sections: \n",text.WriteCellOpts(cell.FgColor(cell.ColorYellow))); err != nil {
		panic(err)
	}
	d.displayPairs(stat.TopSections)

	if err := d.statDisplay.Write("Top sections: \n",text.WriteCellOpts(cell.FgColor(cell.ColorYellow))); err != nil {
		panic(err)
	}
	d.displayPairs(stat.TopMethods)

	if err := d.statDisplay.Write("Top status: \n",text.WriteCellOpts(cell.FgColor(cell.ColorYellow))); err != nil {
		panic(err)
	}
	d.displayPairs(stat.TopStatus)

	if err := d.statDisplay.Write(fmt.Sprintf("Number of requests: "),text.WriteCellOpts(cell.FgColor(cell.ColorYellow))); err != nil {
		panic(err)
	}
	if err := d.statDisplay.Write(fmt.Sprintf("%d\n", stat.NumRequests)); err != nil {
		panic(err)
	}

	if err := d.statDisplay.Write(fmt.Sprintf("Number of bytes: "),text.WriteCellOpts(cell.FgColor(cell.ColorYellow))); err != nil {
		panic(err)
	}
	if err := d.statDisplay.Write(fmt.Sprintf("%d\n", stat.BytesCount)); err != nil {
		panic(err)
	}

}

func fmtDuration(d time.Duration) string {
	uptime := int(d)/1000000000
	s := uptime % 60
	m := uptime / 60 % 60
	h := uptime / 3600
	return fmt.Sprintf("%02dh%02dmin%02ds", h, m, s)
}


func (d *Displayer) update(ctx context.Context) {
	startTime := time.Now().Round(time.Second)
	ticker := time.NewTicker(time.Second)
	for {
		select {

		case <-ticker.C:
			d.uptimeDisplay.Reset()
			if err := d.uptimeDisplay.Write(fmt.Sprintf("%s", fmtDuration(time.Since(startTime)))); err != nil {
				panic(err)
			}
		case info := <-d.StatChan:

			if err := d.histogram.Add([]int{info.NumRequests}); err != nil {
				panic(err)
			}
			d.statDisplay.Reset()
			d.displayInfo(info)

		case alert := <-d.AlertChan:
			d.alertDisplay.Reset()
			if alert.Alert {
				if err := d.alertDisplay.Write(fmt.Sprintf("\n High traffic generated an alert - hits = %d, triggered at %s\n", alert.NumTraffic, time.Now().Format("15:04:05, January 02 2006")), text.WriteCellOpts(cell.FgColor(cell.ColorRed))); err != nil {
					panic(err)
				}
			} else {
				if err := d.alertDisplay.Write(fmt.Sprintf("High traffic has recovered, triggered at %s\n", time.Now().Format("15:04:05, January 02 2006")), text.WriteCellOpts(cell.FgColor(cell.ColorGreen))); err != nil {
					panic(err)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (d *Displayer) Run(){
	go d.update(d.ctx)
	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			d.cancel()
		}
	}
	container.SplitPercent(40)
	box, err := termbox.New()

	container , err := container.New(
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


	defer box.Close()

	if err != nil {
		panic(err)
	}

	if err := termdash.Run(d.ctx, box, container, termdash.KeyboardSubscriber(quitter)); err != nil {
		panic(err)
	}
}
