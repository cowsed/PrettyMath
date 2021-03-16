package attractor2d

import (
	"fmt"

	"github.com/AllenDang/giu"
	"math/rand"
)

var availableAnimationTitles []string = []string{"Random Walk", "Parameter Functions", "BuildUp"}

var baseRandomWalker = RandomWalkerInit()
var availableAnimations []animationMaker = []animationMaker{&baseRandomWalker, &parameterAlter{}, &buildUp{5}}

type animationMaker interface {
	makeFrames(frameAmt int, outPath string, r renderer) //Makes the frame of the video
	makeSetup() giu.Widget
	deepCopy() animationMaker
}

//Random Walks through parameter space
type randomWalker struct {
	labels   []string
	enableds []bool
	stepSize float32
}

//RandomWalkerInit creates a default Random Walker
func RandomWalkerInit() randomWalker {
	return randomWalker{
		labels:   []string{"A", "B", "C", "D", "x0", "y0"},
		enableds: []bool{true, true, true, true, true, true},
		stepSize: 0.1,
	}
}
func (rw *randomWalker) makeSetup() giu.Widget {
	dumbInterface := make([]interface{}, len(rw.enableds))

	checkBoxes := giu.RangeBuilder(
		"Enableds",
		dumbInterface,
		func(i int, _ interface{}) giu.Widget { return giu.Checkbox(rw.labels[i], &rw.enableds[i]) },
	)

	w := giu.Group().Layout(
		giu.Label("Random Walker"),
		giu.InputFloat("Step Size", &rw.stepSize),
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

	return &randomWalker{labels, enableds, rw.stepSize}
}
func (rw *randomWalker) makeFrames(frameAmt int, outPath string, r renderer) {
	for f := 0; f < frameAmt; f++ {
		fmt.Printf("Making frame %d of %d (random walker)\n", f, frameAmt)
		//Set Parameters
		fpath := fmt.Sprintf("%sout%06d.png", outPath, f)
		r.path = fpath
		r.render()
		var newa,newb,newc,newd,newx,newy float64
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
		
		
		fmt.Println("a: %.3f, b: %.3f, c: %.3f, d: %.3f", newa, newb, newc, newd)
		r.vars["a"] = newa
		r.vars["b"] = newb
		r.vars["c"] = newc
		r.vars["d"] = newd			
		r.vars["x"] = newx
		r.vars["y"] = newy
	}
}

//parameterAlter Accepts functions to add on to parameters for example a+=sin(t) where t is the frame of the video / the total number of frames
type parameterAlter struct {
	//[]Expression Elements
}

func (pa *parameterAlter) makeFrames(frameAmt int, outPath string, r renderer) {
	for f := 0; f < frameAmt; f++ {
		fmt.Printf("Making frame %d of %d (Parameter Alterer)", f, frameAmt)

	}
}
func (pa *parameterAlter) makeSetup() giu.Widget {
	return giu.Label("Paramter Alter")
}
func (pa *parameterAlter) deepCopy() animationMaker {
	return &parameterAlter{}
}

//build up makes an animation of 1 frame by graphing some amount of points per frame
type buildUp struct {
	percentPerFrame float32
}

func (bu *buildUp) makeFrames(frameAmt int, outPath string, r renderer) {
	startNumPoints := float64(r.numPoints)
	for f := 0; f < frameAmt; f++ {
		fmt.Printf("Making frame %d of %d (buildup)\n", f, frameAmt)
		tempNumPoints := startNumPoints * (float64(f) / float64(frameAmt))
		r.numPoints = int(tempNumPoints)
		fpath := fmt.Sprintf("%sout%06d.png", outPath, f)
		r.path = fpath
		r.render()
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
