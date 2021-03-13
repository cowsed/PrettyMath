package main

import (
	a2d "./Workspaces/Attractor2D"
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
		g.Button("New 2D Attractor").OnClick(c.SetWorkspace),
	).Build()

}

//SetWorkspace opens the selected workspace (for now the only workspace
func (c *NewWorkspace) SetWorkspace() {
	a := a2d.Init(onClose)
	CurrentWorkspace = &a
}

//The closing function that gets passed to workspaces to remove themselves and replace themselves with the creation tab
func onClose() {
	CurrentWorkspace = &NewWorkspace{}
}
