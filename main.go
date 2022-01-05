package main

import (
	"bufio"
	"fmt"
	g "github.com/AllenDang/giu"
	fifo "github.com/foize/go.fifo"
	"github.com/sqweek/dialog"
	"github.com/valyala/fasthttp"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var namearray []string
var names = fifo.NewQueue()
var arenamesempty = true
var routines int32
var output = "Waiting for user...\n"
var showWindow2 = false
var lineTicks []g.PlotTicker
var xmx = 1.0
var ymx = 1.0
var linedata []float64
var totalchecks = 0
var (
	client = &fasthttp.Client{
		MaxConnsPerHost: 25000,
		Dial: func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, time.Second*3)
		},
	}
)

func callcheck() {
	var i int32 = 0
	go iterategraph()
	for i < routines {
		go check()
		i++
	}
}

func iterategraph() {
	var i = 1
	for true {
		output = "[] r/s: " + strconv.Itoa(totalchecks/i)
		lineTicks = append(lineTicks, g.PlotTicker{Position: float64(i), Label: ""})
		linedata = append(linedata, float64(totalchecks/i))
		xmx = float64(i)
		if float64(totalchecks/i) > ymx {
			ymx = float64(totalchecks / i)
		}
		time.Sleep(1000 * time.Millisecond)
		i++
	}
}

func callexit() {
	output += "-------------------\nShutting down...\n"
	go exit()
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func check() {
	var req = fasthttp.AcquireRequest()
	var resp = fasthttp.AcquireResponse()
	req.Header.SetMethod("GET")
	for true {
		var user = fmt.Sprint(names.Next())
		names.Add(user)
		req.SetRequestURI("https://letterboxd.com/s/checkusername?q=" + user)
		err := client.Do(req, resp)
		if err != nil {
			return
		}
		if strings.Contains(string(resp.Body()), "AVAILABLE") {
			//output += "User " + user + " is available!\n"
		} else if strings.Contains(string(resp.Body()), "TAKEN") || strings.Contains(string(resp.Body()), "ILLEGAL") || strings.Contains(string(resp.Body()), "INVALID") {
			//output += "User " + user + " is not available.\n"
		} else {
			//output += "Error when checking " + user + ".\n"
		}
		totalchecks++
	}
}

func opennames() {
	filename, _ := dialog.File().Filter("Username List (.txt)", "txt").Load()
	namearray, _ = readLines(filename)
	for _, name := range namearray {
		names.Add(name)
	}
	output += "Successfully imported list of " + strconv.FormatInt(int64(names.Len()), 10) + " usernames."
	arenamesempty = false
}

func exit() {
	exit()
}

func togglestats() {
	if showWindow2 {
		showWindow2 = false
		window2x = 405
		wnd.SetSize(400, 300)
	} else {
		showWindow2 = true
		window2x = 400
		wnd.SetSize(900, 300)
	}
}

//not currently working to fix overlap
var window2x = 400
var window2y = 0

func loop() {
	g.Window("Command Center").Flags(g.WindowFlagsNoResize).Flags(g.WindowFlagsNoMove).Pos(0, 0).Size(400, 300).Layout(
		g.Label("instant - go"),
		g.Separator(),
		g.InputTextMultiline(&output).Size(-1, 220).Flags(g.InputTextFlagsReadOnly),
		g.Row(
			g.Button("Check").OnClick(callcheck).Disabled(arenamesempty),
			g.Button("Open Names").OnClick(opennames),
			g.Button("Exit").OnClick(callexit),
			g.Button("Toggle Stats").OnClick(togglestats),
			g.InputInt(&routines).Label("Routines").Size(g.Auto),
		),
	)
	if showWindow2 {
		g.Window("Stats").IsOpen(&showWindow2).Flags(g.WindowFlagsNoResize).Flags(g.WindowFlagsNoMove).Flags(g.WindowFlagsAlwaysUseWindowPadding).Pos(float32(window2x), float32(window2y)).Size(500, 300).Flags(g.WindowFlagsNoMove).Layout(
			g.Plot("requests").AxisLimits(0, xmx, 0, ymx, g.ConditionOnce).XTicks(lineTicks, false).Plots(
				g.PlotLine("r/s line", linedata),
			).XAxeFlags(g.PlotAxisFlagsAutoFit).YAxeFlags(g.PlotAxisFlagsAutoFit, g.PlotAxisFlagsAutoFit, g.PlotAxisFlagsAutoFit),
		)
	}

}

//create window globally so it may be accessed globally
var wnd = g.NewMasterWindow("instant. ---", 400, 300, g.MasterWindowFlagsNotResizable)

func main() {
	g.SetDefaultFont("Consola", 12)
	wnd.Run(loop)
}
