package main

import (
	"fmt"

	g "github.com/AllenDang/giu"
	workspace "github.com/cowsed/PrettyMath/Workspaces"
)

var statusWindowShown = false

var dummyReciever = workspace.ProgressUpdate{Status: "Dummy", Amount: 0.75}
var statuses []workspace.ProgressUpdate = []workspace.ProgressUpdate{}
var communicators []chan workspace.ProgressUpdate = []chan workspace.ProgressUpdate{}

//AddProcess creates a new process status receiver and adds it to the slices that keep track
func AddProcess() chan workspace.ProgressUpdate {
	//Create communication channel
	//TODO tell  user to stop if theres a lot going >= number of cores on machine
	//Buffer size of 5 just cuz
	c := make(chan workspace.ProgressUpdate, 5)
	communicators = append(communicators, c)
	statuses = append(statuses, workspace.ProgressUpdate{Status: "No Info Yet", Description: "No description", Amount: .0})
	return c
}

func queryComms() {
	doUpdate := false
	for i, ch := range communicators {
		select {
		case status, ok := <-ch:
			if ok {
				statuses[i] = workspace.ProgressUpdate{Status: status.Status, Description: status.Description, Amount: status.Amount}
				//There was a change so update
				doUpdate = true
				//IF value is end close channel and remove it
			} else {
				defer func() {
					communicators = removeComm(communicators, i)
					statuses = removeStatus(statuses, i)
					//This causes problems because it will try to access at i but if the previous iteration removed something
					//it is now out of range
					//OOh, I know. Defer
				}()
				doUpdate = true
			}
		default:
			//No Value just keep going
		}
	}
	if doUpdate {
		g.Update()
	}
}

//ToggleStatusWindow Toggles the visibility of the status window
func ToggleStatusWindow() {
	statusWindowShown = !statusWindowShown

}

//Build the whole window
func buildStatusWindow() {
	var content g.Widget
	fmt.Println("comms", communicators)

	if len(communicators) == 0 {
		content = g.Label("No Running processes")
	}
	g.Window("Statuses").IsOpen(&statusWindowShown).Pos(400, 60).Layout(
		content,
		BuildAllStatuses(),
	)
}

//BuildAllStatuses builds all of the statuses known for imgui
func BuildAllStatuses() g.Widget {

	//needed for rangebuilder
	dumbInterface := make([]interface{}, len(statuses))
	widget := g.RangeBuilder("ListOfStatuses", dumbInterface,
		func(i int, v interface{}) g.Widget {
			r := statuses[i]
			w := g.Group().Layout(
				g.Label(r.Status),
				g.ProgressBar(float32(r.Amount)),
				g.Tooltip(r.Description),
			)
			return w
		})
	return widget
}

//Helper functions for dealing with workers
func removeStatus(s []workspace.ProgressUpdate, i int) []workspace.ProgressUpdate {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
func removeComm(s []chan workspace.ProgressUpdate, i int) []chan workspace.ProgressUpdate {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
