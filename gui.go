package main

import (
	g "github.com/AllenDang/giu"

	"./Workspaces/Attractor2D"
)

var NewWindowOpen = true
var Attractor2DOpen = true
//var Workspaces []g.Widget
var ws attractor2D.Attractor2DWorkspace

func loop() {

	g.SingleWindow("Images").Layout(
		g.TabBar("TabBar").Layout(
			&ws,
			g.TabItem("+").Layout(
				g.Label("Shouldnt have closed that other window, huh bud"),
			),
		),
	)

}

func main() {
	//colGradient.Init()
	ws=attractor2D.Init()
	//a2dWS := attractor2D.Init()
	//Workspaces = append(Workspaces, &a2dWS)

	wnd := g.NewMasterWindow("Fun Graphics Stuff", 1200, 800, 0, nil)

	//loadImage()

	wnd.Run(loop)

}
