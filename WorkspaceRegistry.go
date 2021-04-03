package main

import (
	plot "github.com/cowsed/PrettyMath/Workspaces/2DPlotter"
	a2d "github.com/cowsed/PrettyMath/Workspaces/Attractor2D"
	ocl "github.com/cowsed/PrettyMath/Workspaces/OpenCL"
	g "github.com/AllenDang/giu"
	
	"runtime"
)

//NewWorkspace is an empty struct (to fulfill g.Widget)for a page that allows the creation of workspaces
type NewWorkspace struct {
}

//Build is the immediate mode GUI Build function
func (c *NewWorkspace) Build() {
	//Build the creation tab
	g.TabItem("Create New").Layout(
		g.Label("Create New: "),
		g.Button("New 2D Attractor").OnClick(c.SetA2D),
		g.Button("New Plot").OnClick(c.SetPlot),
		g.Button("New OpenCL").OnClick(c.SetOpenCL),
	).Build()
}

//SetA2D opens the attractor 2d workspace
func (c *NewWorkspace) SetA2D() {
	a := a2d.Init(onClose, AddProcess)
	CurrentWorkspace = &a
}

//SetPlot opens the plotting workspace
func (c *NewWorkspace) SetPlot() {
	p := plot.Init(onClose)
	CurrentWorkspace = &p
}

//SetPlot opens the plotting workspace
func (c *NewWorkspace) SetOpenCL() {
	p := ocl.Init(onClose)
	CurrentWorkspace = &p
}

//The closing function that gets passed to workspaces to remove themselves and replace themselves with the creation tab
func onClose() {
	CurrentWorkspace = &NewWorkspace{}
	runtime.GC()
}
