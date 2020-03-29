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
	"github.com/mum4k/termdash/widgets/text"
	"log-monitor/monitoring"
)

type Displayer struct {
	StatChan chan monitoring.StatRecord
	AlertChan chan string
	statDisplay *text.Text
	alertDisplay *text.Text
	container *container.Container
	box *termbox.Terminal
	ctx context.Context
	cancel context.CancelFunc
}



func (d *Displayer) Init(statChan chan monitoring.StatRecord, alertChan chan string, ctx context.Context, cancel context.CancelFunc){
	d.StatChan = statChan
	d.AlertChan = alertChan
	d.ctx = ctx
	d.cancel = cancel
	statDisplay, err := text.New(text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	d.statDisplay = statDisplay

	alertDisplay, err := text.New(text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	d.alertDisplay = alertDisplay

	d.box, err = termbox.New()
	if err != nil {
		panic(err)
	}

	d.container , err = container.New(
		d.box,
		container.Border(linestyle.Light),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.SplitVertical(
			container.Left(
				container.Border(linestyle.Light),
				container.BorderTitle("Traffic info"),
				container.PlaceWidget(statDisplay)),
			container.Right(
				container.Border(linestyle.Light),
				container.BorderTitle("Alerts"),
				container.PlaceWidget(alertDisplay),
			),
		),
	)
	if err != nil {
		panic(err)
	}

}

func (d *Displayer) displayPairs( pairs []monitoring.Pair) {
	for _, pair := range pairs{
		err := d.statDisplay.Write(fmt.Sprintf("    %s: %d\n", pair.Key, pair.Value))
		if err != nil {
			panic(err)
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

	//out += fmt.Sprintf("\nNumber of requests: %d", stat.NumRequests)
	//out += fmt.Sprintf("Number of bytes: %d", stat.BytesCount )
}


func (d *Displayer) update(ctx context.Context) {

	for {
		select {
		case in := <-d.StatChan:
			d.statDisplay.Reset()
			d.displayInfo(in)

		case in := <-d.AlertChan:
			d.statDisplay.Reset()
			if err := d.alertDisplay.Write(fmt.Sprintf("%s\n", in, ), text.WriteCellOpts(cell.FgColor(cell.ColorRed))); err != nil {
				panic(err)
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
	if err := termdash.Run(d.ctx, d.box, d.container, termdash.KeyboardSubscriber(quitter)); err != nil {
		panic(err)
	}
}
