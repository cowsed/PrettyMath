package main

import (
	"./Workspaces"
	"fmt"
	g "github.com/AllenDang/giu"
)

var StatusWindowShown = false
var dummyReciever = workspace.ProgressUpdate{Status: "Dummy", Amount: 0.75}
var statuses []workspace.ProgressUpdate = []workspace.ProgressUpdate{}
var communicators []chan workspace.ProgressUpdate = []chan workspace.ProgressUpdate{}

//Having to buffer the channels probably isnt the best idea but since
func AddProcess() chan workspace.ProgressUpdate {
	//Create communication channel
	//TODO tell  user to stop if theres a lot going >= number of cores on machine
	//Buffer size of 5 just cuz
	c := make(chan workspace.ProgressUpdate, 5)
	communicators = append(communicators, c)
	statuses = append(statuses, workspace.ProgressUpdate{"No Info Yet", 0.0})
	return c
}

func removeStatus(s []workspace.ProgressUpdate, i int) []workspace.ProgressUpdate {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
func removeComm(s []chan workspace.ProgressUpdate, i int) []chan workspace.ProgressUpdate {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}


func queryComms() {
	doUpdate := false
	for i, ch := range communicators {
		select {
		case status, ok := <-ch:
			if ok {
				println("Valuw was read. -- update statuses")
				statuses[i] = workspace.ProgressUpdate{status.Status, status.Amount}
				//There was a change so update
				doUpdate = true
				//IF value is end close channel and remove it
			} else {
				println("Channel closed! -- remove from list")
				fmt.Println("Slice: ", statuses, "index", i)
				communicators = removeComm(communicators, i)
				statuses = removeStatus(statuses, i)
			}
		default:
			println("No value ready, moving on.")
		}
	}
	if doUpdate {
		g.Update()
	}

}

//ToggleStatusWindow Toggles the visibility of the status window
func ToggleStatusWindow() {
	StatusWindowShown = !StatusWindowShown
}

//Build the whole window
func buildStatusWindow() {
	println("building status window")
	g.Window("Statuses").IsOpen(&StatusWindowShown).Pos(400, 60).Layout(
		g.Label("No Running processes"),
		//&dummyReciever, //.makeInfo(),
		BuildAllStatuses(),
	)
}

//Build all of the statuses known
func BuildAllStatuses() g.Widget {

	//needed for rangebuilder
	dumbInterface := make([]interface{}, len(statuses))
	widget := g.RangeBuilder("ListOfStatuses", dumbInterface,
		func(i int, v interface{}) g.Widget {
			r := statuses[i]
			w := g.Group().Layout(
				g.Label(r.Status),
				g.ProgressBar(float32(r.Amount)),
			)
			return w
		})
	return widget
}
