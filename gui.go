package main

import (
	g "github.com/AllenDang/giu"
)

//CurrentWorkspace is the workspace that will be rendered (initial value is the creation page)
var CurrentWorkspace g.Widget = &NewWorkspace{}

//Create Top Level containers
func loop() {

	g.SingleWindow("Images").Layout(
		g.TabBar("TabBar").Layout(
			CurrentWorkspace,
		),
	)

}

func main() {
	//Create Window
	wnd := g.NewMasterWindow("PrettyMath", 1200, 800, 0, nil)
	//Run it
	wnd.Run(loop)

}
