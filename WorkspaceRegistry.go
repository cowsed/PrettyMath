package main

import (
	"runtime"

	workspace "github.com/cowsed/PrettyMath/Workspaces"
	_ "github.com/cowsed/PrettyMath/Workspaces/2DPlotter"
	_ "github.com/cowsed/PrettyMath/Workspaces/Attractor2D"
	_ "github.com/cowsed/PrettyMath/Workspaces/ModelViewer"

	// _ "github.com/cowsed/PrettyMath/Workspaces/OpenCL" //Doesn't work currently

	ogl "github.com/cowsed/PrettyMath/Workspaces/OpenGL"

	g "github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
)

//NewWorkspace is an empty struct (to fulfill g.Widget)for a page that allows the creation of workspaces
type NewWorkspace struct {
}

//Build builds the page full of options for workspaces to create
func (NWS *NewWorkspace) Build() {
	//Build the creation tab
	g.TabItem("Create New").Layout(
		g.Label("New Page: "),
		g.Custom(func() {
			for _, ws := range workspace.RegisteredWorkspaces {
				if imgui.Button("Create " + ws.Name) {
					CurrentWorkspaces = append(CurrentWorkspaces, ws.InitFunc(onClose, AddProcess))
				}
			}
		}),
		g.Button("New OpenGL").OnClick(NWS.SetOpenGL),
	).Build()
}

//SetPlot opens the plotting workspace
func (c *NewWorkspace) SetOpenGL() {
	p := ogl.Init(onClose)
	CurrentWorkspaces = append(CurrentWorkspaces, &p)
}

//The closing function that gets passed to workspaces to remove themselves and replace themselves with the creation tab
func onClose() {
	CurrentWorkspaces = []g.Widget{&NewWorkspace{}}

	runtime.GC()
}
