package main

import (
	"time"
	g "github.com/AllenDang/giu"
)

//CurrentWorkspace is the workspace that will be rendered (initial value is the creation page)
var CurrentWorkspace g.Widget = &NewWorkspace{}

var wnd *g.MasterWindow


//Create Top Level containers
func loop() {
	
	statusBar := g.Group().Layout(
		g.ArrowButton("Open Statuses", g.DirectionDown).OnClick(ToggleStatusWindow),
		g.Tooltip("Press to open status window"),
	)

	fullsizex,fullsizey:=wnd.GetSize()
	//TODO: Figure out why setting position and size doesnt work
	g.SingleWindow("Images").Flags(g.WindowFlagsNoBringToFrontOnFocus+g.WindowFlagsNoDecoration).Pos(0,0).Size(float32(fullsizex),float32(fullsizey)).Layout(
		g.Line(
			statusBar,
			g.TabBar("TabBar").Layout(
				CurrentWorkspace,
			),
		),
	)
	if statusWindowShown {
		buildStatusWindow()
	}

}

func main() {
	//Query the comms etc and forever
	go func(){
		for {
			queryComms()
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	//Create Window
	wnd = g.NewMasterWindow("PrettyMath", 1200, 800, 0, nil)
	//Run it
	wnd.Run(loop)

}
