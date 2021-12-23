package attractor2d

import (
	"fmt"
	"image/png"
	"os"

	workspace "github.com/cowsed/PrettyMath/Workspaces"

	"math/rand"

	"github.com/AllenDang/giu"
)

var availableAnimationTitles []string = []string{"Random Walk", "Parameter Functions", "BuildUp"}

var baseRandomWalker = randomWalkerInit()
var availableAnimations []animationMaker = []animationMaker{&baseRandomWalker, &parameterAlter{}, &buildUp{5}}

type animationMaker interface {
	makeFrames(frameAmt int, outPath string, r renderer, processCreator func() chan workspace.ProgressUpdate) //Makes the frame of the video
	makeSetup() giu.Widget
	deepCopy() animationMaker
}

//Random Walks through parameter space
type randomWalker struct {
	skipBlanks        bool
	skipBlanksPercent float32
	labels            []string
	enableds          []bool
	stepSize          float32
}

//RandomWalkerInit creates a default Random Walker
func randomWalkerInit() randomWalker {
	return randomWalker{
		skipBlanks:        false,
		skipBlanksPercent: 0.3,
		labels:            []string{"A", "B", "C", "D", "x0", "y0"},
		enableds:          []bool{true, true, true, true, true, true},
		stepSize:          0.1,
	}
}
func (rw *randomWalker) makeSetup() giu.Widget {
	dumbInterface := make([]interface{}, len(rw.enableds))

	checkBoxes := giu.RangeBuilder(
		"Enableds",
		dumbInterface,
		func(i int, _ interface{}) giu.Widget { return giu.Checkbox(rw.labels[i], &rw.enableds[i]) },
	)

	var blankSkipper giu.Widget
	if rw.skipBlanks {
		blankSkipper = giu.Group().Layout(

			//if rw.skipBlanks{
			giu.InputFloat("Skip Blank Percent", &rw.skipBlanksPercent).Size(80),
			//}
		)
	}
	w := giu.Group().Layout(
		giu.Label("Random Walker"),
		giu.InputFloat("Step Size", &rw.stepSize),
		giu.Label("Skip Blanks"), giu.Tooltip("if true skips any frame where individual points drawn / max image points < maxImagePoints.skipBlanksPercent"),
		giu.Checkbox("Skip Blanks", &rw.skipBlanks),
		blankSkipper,
		giu.Label("Enabled"),
		giu.Tooltip("Which parameters an be effected by the random walk"),
		checkBoxes,
	)

	return w
}
func (rw *randomWalker) deepCopy() animationMaker {
	labels := make([]string, len(rw.labels))
	copy(labels, rw.labels)

	enableds := make([]bool, len(rw.enableds))
	copy(enableds, rw.enableds)

	return &randomWalker{false, 0.3, labels, enableds, rw.stepSize}
}
func (rw *randomWalker) makeFrames(frameAmt int, outPath string, r renderer, processCommCreator func() chan workspace.ProgressUpdate) {

	processDescription := fmt.Sprintf("Renders %d frames of a random walk through parameter space.\nBegan at {time}", frameAmt)
	communicator := processCommCreator()
	//mabye defer close(communicator)
	go func() {
		println("Making animation - goroutine")

		for f := 0; f < frameAmt; f++ {
			println("Making animation - loop - goroutine")

			amt := float64(f) / float64(frameAmt)
			communicator <- workspace.ProgressUpdate{
				Status:      "Working",
				Description: processDescription,
				Amount:      amt,
			}

			//Set Parameters
			fpath := fmt.Sprintf("%sout%06d.png", outPath, f)
			img := r.render()

			//Save image to file
			f, _ := os.Create(fpath)
			png.Encode(f, img)

			var newa, newb, newc, newd, newx, newy float64
			//Set up the next round
			if rw.enableds[0] {
				newa = r.vars["a"] + ((rand.Float64()*2)-1)*float64(rw.stepSize)
			}
			if rw.enableds[1] {
				newb = r.vars["b"] + ((rand.Float64()*2)-1)*float64(rw.stepSize)
			}
			if rw.enableds[2] {
				newc = r.vars["d"] + ((rand.Float64()*2)-1)*float64(rw.stepSize)
			}
			if rw.enableds[3] {
				newd = r.vars["d"] + ((rand.Float64()*2)-1)*float64(rw.stepSize)
			}
			if rw.enableds[4] {
				newx = r.vars["z"] + ((rand.Float64()*2)-1)*float64(rw.stepSize)
			}
			if rw.enableds[5] {
				newy = r.vars["y"] + ((rand.Float64()*2)-1)*float64(rw.stepSize)
			}

			fmt.Printf("a: %.3f, b: %.3f, c: %.3f, d: %.3f\n", newa, newb, newc, newd)
			r.vars["a"] = newa
			r.vars["b"] = newb
			r.vars["c"] = newc
			r.vars["d"] = newd
			r.vars["x"] = newx
			r.vars["y"] = newy
		}
		communicator <- workspace.ProgressUpdate{Status: "End", Description: processDescription, Amount: .0}
		//Notify user of finishing
		fmt.Println("\n\n\n ---- Finished animation ----")
		close(communicator)
	}()
}

//parameterAlter Accepts functions to add on to parameters for example a+=sin(t) where t is the frame of the video / the total number of frames
type parameterAlter struct {
	//[]Expression Elements
}

//For now, doesnt do anything
func (pa *parameterAlter) makeFrames(frameAmt int, outPath string, r renderer, processCreator func() chan workspace.ProgressUpdate) {
	for f := 0; f < frameAmt; f++ {
		fmt.Printf("Making frame %d of %d (Parameter Alterer)", f, frameAmt)

	}
}
func (pa *parameterAlter) makeSetup() giu.Widget {
	return giu.Label("Paramter Alter\nNot implemented yet")
}
func (pa *parameterAlter) deepCopy() animationMaker {
	return &parameterAlter{}
}

//build up makes an animation of 1 frame by graphing some amount of points per frame
type buildUp struct {
	percentPerFrame float32
}

//For now, Blocking
func (bu *buildUp) makeFrames(frameAmt int, outPath string, r renderer, processCreator func() chan workspace.ProgressUpdate) {
	startNumPoints := float64(r.numPoints)
	for f := 0; f < frameAmt; f++ {
		fmt.Printf("Making frame %d of %d (buildup)\n", f, frameAmt)
		tempNumPoints := startNumPoints * (float64(f) / float64(frameAmt))
		r.numPoints = int(tempNumPoints)
		fpath := fmt.Sprintf("%sout%06d.png", outPath, f)
		img := r.render()

		//Save image to file
		f, _ := os.Create(fpath)
		png.Encode(f, img)

	}
}
func (bu *buildUp) makeSetup() giu.Widget {
	w := giu.Group().Layout(
		giu.Label("Build Up"),
		giu.Label("Over the desired number of frames build the image up in chunks sizes calculated from the number of frames"),
		//giu.InputFloat("Percent per frame(0-100)", &bu.percentPerFrame).Size(80),
	)
	return w
}
func (bu *buildUp) deepCopy() animationMaker {
	return &buildUp{bu.percentPerFrame}
}
