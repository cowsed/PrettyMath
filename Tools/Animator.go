package tools

import (
	"fmt"
	"sort"

	"github.com/AllenDang/giu"
)

//Float32 Animator (needs an editor)
type float32Frame struct {
	pos   float32
	value float32
}
type Float32Animator struct {
	keyFrames []float32Frame
	min, max  float32
}

func (a *Float32Animator) AddFrame(position float32, value float32) {
	frame := float32Frame{position, value}
	a.keyFrames = append(a.keyFrames, frame)
	if value > a.max {
		a.max = value
	} else if value < a.min {
		a.min = value
	}
	a.sort()
}
func (a *Float32Animator) GetValue(pos float32) float32 {
	low := 0
	high := len(a.keyFrames) - 1
	middle := 0
	for low <= high {
		middle = (low + high) / 2
		var frame float32Frame = a.keyFrames[middle]
		if frame.pos > pos {
			high = middle - 1 //search low end of array
		} else if frame.pos < pos {
			low = middle + 1 //search high end of array
		} else {
			return frame.value
		}
	}

	//return interpolated value
	if a.keyFrames[middle].pos > pos {
		middle--
	}
	first := middle
	second := middle + 1

	if second >= len(a.keyFrames) {
		return a.keyFrames[len(a.keyFrames)-1].value
	}
	if first < 0 {
		return a.keyFrames[0].value
	}
	pointFirst := a.keyFrames[first]
	pointSecond := a.keyFrames[second]

	fmt.Println("first", pointFirst, "second", pointSecond)

	return lerp32(pointSecond.value, pointFirst.value, (pos-pointFirst.pos)/(pointSecond.pos-pointFirst.pos))
}
func lerp32(a, b, pos float32) float32 {
	return (a*float32(pos) + b*float32(1-pos))
}

func NewFloat32Animator() Float32Animator {
	a := Float32Animator{}
	a.AddFrame(0, 0)
	a.AddFrame(0.5, 0.)
	a.AddFrame(1, 1)
	//fmt.Println("a",a)
	return a
}
func (a *Float32Animator) sort() {
	sort.SliceStable(a.keyFrames, func(i, j int) bool { return a.keyFrames[i].pos < a.keyFrames[j].pos })
}

type Float32AnimatorWidget struct {
	animator *Float32Animator
}

func NewFloat32AnimatorWidget(animator *Float32Animator) Float32AnimatorWidget {
	return Float32AnimatorWidget{animator}
}

func (aw Float32AnimatorWidget) Build() {
	data := make([]float64, 11)
	xs := make([]float64, 11)

	for i := 0; i < 11; i++ {
		fmt.Println("Getting at", float32(i)*.1, " got ", aw.animator.GetValue(float32(i)*.1))
		data[i] = float64(aw.animator.GetValue(float32(i) * .1))
		xs[i] = float64(i) * .1
	}
	giu.Group().Layout(
		giu.Label("I do things"),
		giu.Plot("Preview").Plots(
			giu.PlotLineXY("PreviewData", xs, data),
		).AxisLimits(0, 1, float64(aw.animator.min)-.1, float64(aw.animator.max)+.1, giu.ConditionAlways),
	).Build()
}
