package main

import (
	a2d "./Workspaces/Attractor2D"
	plot "./Workspaces/2DPlotter"
	g "github.com/AllenDang/giu"
)

//NewWorkspace is an empty struct (to fulfill g.Widget)for a page that allows the creation of workspaces
type NewWorkspace struct {
}

//Build is the immediate mode GUI Build function
func (c *NewWorkspace) Build() {
	//Build the creation tab
	g.TabItem("+").Layout(
		g.Label("Create New: "),
		g.Button("New 2D Attractor").OnClick(c.SetA2D),
		g.Button("New Plot").OnClick(c.SetPlot),

).Build()

}

//SetA2D opens the attractor 2d workspace
func (c *NewWorkspace) SetA2D() {
	a := a2d.Init(onClose)
	CurrentWorkspace = &a
}
//SetPlot opens the plotting workspace
func (c *NewWorkspace) SetPlot() {
	p := plot.Init(onClose)
	CurrentWorkspace = &p
}


//The closing function that gets passed to workspaces to remove themselves and replace themselves with the creation tab
func onClose() {
	CurrentWorkspace = &NewWorkspace{}
}
