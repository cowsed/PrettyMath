package main

import (
	"runtime"

	plot "github.com/cowsed/PrettyMath/Workspaces/2DPlotter"
	a2d "github.com/cowsed/PrettyMath/Workspaces/Attractor2D"
	ocl "github.com/cowsed/PrettyMath/Workspaces/OpenCL"
	ogl "github.com/cowsed/PrettyMath/Workspaces/OpenGL"

	g "github.com/AllenDang/giu"
	modelviewer "github.com/cowsed/PrettyMath/Workspaces/ModelViewer"
)

//NewWorkspace is an empty struct (to fulfill g.Widget)for a page that allows the creation of workspaces
type NewWorkspace struct {
}

//Build is the immediate mode GUI Build function
func (NWS *NewWorkspace) Build() {
	//Build the creation tab
	g.TabItem("Create New").Layout(
		g.Label("Create New: "),
		g.Button("New 2D Attractor").OnClick(NWS.SetA2D),
		g.Button("New Plot").OnClick(NWS.SetPlot),
		g.Button("New OpenCL").OnClick(NWS.SetOpenCL),
		g.Button("New OpenGL").OnClick(NWS.SetOpenGL),
		g.Button("New Model Viewer").OnClick(NWS.SetModelViewer),
	).Build()
}

//SetModelViewer opens the attractor 2d workspace
func (c *NewWorkspace) SetModelViewer() {
	a := modelviewer.Init(onClose)
	CurrentWorkspace = &a
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

//SetPlot opens the plotting workspace
func (c *NewWorkspace) SetOpenGL() {
	p := ogl.Init(onClose)
	CurrentWorkspace = &p
}

//The closing function that gets passed to workspaces to remove themselves and replace themselves with the creation tab
func onClose() {
	CurrentWorkspace = &NewWorkspace{}
	runtime.GC()
}
