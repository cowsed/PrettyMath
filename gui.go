package main

import (
	"time"

	"github.com/AllenDang/giu/imgui"
	ogl "github.com/cowsed/PrettyMath/Workspaces/OpenGL"

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

	fullsizex, fullsizey := wnd.GetSize()

	imgui.SetNextWindowSize(imgui.Vec2{X: float32(fullsizex), Y: float32(fullsizey)})
	imgui.SetNextWindowPos(imgui.Vec2{X: 0, Y: 0})
	g.SingleWindow("Images").Layout(
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
	go func() {
		for {
			queryComms()
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	//Create Window
	wnd = g.NewMasterWindow("PrettyMath", 1200, 800, 0, nil)

	var start = (ogl.Init(onClose))
	CurrentWorkspace = &start

	//Run it
	wnd.Run(loop)

}
