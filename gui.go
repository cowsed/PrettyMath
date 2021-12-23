package main

import (
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/AllenDang/giu/imgui"

	g "github.com/AllenDang/giu"
)

//CurrentWorkspace is the workspace that will be rendered (initial value is the creation page)
//var CurrentWorkspace g.Widget = &NewWorkspace{}
var CurrentWorkspaces []g.Widget = []g.Widget{&NewWorkspace{}}

var wnd *g.MasterWindow

//Create Top Level containers
func loop() {
	statusBar := g.Group().Layout(
		g.ArrowButton("Open Statuses", g.DirectionDown).OnClick(ToggleStatusWindow),
		g.Tooltip("Press to open status window"),
	)

	fullsizex, fullsizey := wnd.GetSize()

	imgui.SetNextWindowSize(imgui.Vec2{X: float32(fullsizex), Y: float32(fullsizey)})
	imgui.SetNextWindowPos(imgui.Vec2{X: 0, Y: 0})
	g.SingleWindow("Images").Layout(
		g.Line(
			statusBar,
			g.TabBar("TabBar").Layout(
				CurrentWorkspaces...,
			).Flags(g.TabBarFlagsAutoSelectNewTabs),
		),
	)
	if statusWindowShown {
		buildStatusWindow()
	}

}

func refresh() {
	ticker := time.NewTicker(time.Millisecond * 32)

	for {
		g.Update()

		<-ticker.C
	}
}

func main() {
	f, err := os.Create("cpuprofile.prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer f.Close() // error handling omitted for example
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	//Query the comms etc and forever
	go func() {
		for {
			queryComms()
			time.Sleep(1000 * time.Millisecond)
		}
	}()
	//Keep the window refreshing
	go refresh()

	//Create Window
	wnd = g.NewMasterWindow("PrettyMath", 1200, 800, 0, nil)

	//Run it
	wnd.Run(loop)

}
