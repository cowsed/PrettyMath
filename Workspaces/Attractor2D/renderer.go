package attractor2d

import (
	"fmt"
	"image"
	"image/color"
	"math"

	ep "../../ExpressionParser"
	"../../Tools"
)

//asyncRenderer holds all the parameters needed to render asynchornously without carrying the extra bagage of all the gui stuff
type renderer struct {
	vars        map[string]float64
	xExpression ep.ExpressionElement
	yExpression ep.ExpressionElement

	nthRoot       float64
	gradient      tools.Gradient
	width, height int
	offx, offy    int
	scaleFactor   float64
	numPoints     int
}

func (r *renderer) render() *image.RGBA {
	//Setup
	pointMap := make(map[[2]int]int)
	var maxPointsPerCell = 1
	var drawnPoints = 0
	var drawnPointsIndividual = 0

	//Drawing
	for i := 0; i < int(r.numPoints); i++ {
		newx := r.xExpression.BecomeNumber(r.vars)
		newy := r.yExpression.BecomeNumber(r.vars)

		r.vars["x"] = newx
		r.vars["y"] = newy

		disx := int(newx*float64(r.width)*float64(r.scaleFactor) + float64(r.offx))
		disy := int(newy*float64(r.height)*float64(r.scaleFactor) + float64(r.offy))

		//If it is in the bounds of the image
		if disx < int(r.width) && disy < int(r.height) && disx >= 0 && disy >= 0 {
			if inMap([2]int{disx, disy}, pointMap) {
				//if its already there, increment count
				pointMap[[2]int{disx, disy}] = pointMap[[2]int{disx, disy}] + 1
			} else {
				pointMap[[2]int{disx, disy}] = 1
				drawnPointsIndividual++
			}
			drawnPoints++
			maxPointsPerCell = maxI(maxPointsPerCell, pointMap[[2]int{disx, disy}])
		}
	}
	//Render Data
	print("Max Points: ")
	println(maxPointsPerCell)
	print("Points Drawn: ")
	println(drawnPoints)
	print("Individual Points: ")
	fmt.Println(drawnPointsIndividual)
	print("Possible Points: ")
	fmt.Println(r.width * r.height)

	//Making into an image
	upLeft := image.Point{0, 0}
	lowRight := image.Point{int(r.width), int(r.height)}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	for x := 0; x < int(r.width); x++ {
		for y := 0; y < int(r.height); y++ {
			var amt float64 = 0
			if inMap([2]int{x, y}, pointMap) {
				amt = float64(pointMap[[2]int{x, y}]) / float64(maxPointsPerCell)
			} else {
				amt = 0
			}
			amt = math.Pow(amt, 1.0/r.nthRoot)
			newCol := r.gradient.GetColorAt(amt)
			img.Set(x, y, arrToRGBA(newCol))
		}
	}

	return img
}

//Helpful functions
func minI(a, b int) int {
	if a < b {
		return a
	}
	return b

}
func maxI(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func inMap(key [2]int, m map[[2]int]int) bool {
	_, in := m[key]
	return in
}

func arrToRGBA(arr [4]float32) color.RGBA {
	return color.RGBA{uint8(arr[0] * 255), uint8(arr[1] * 255), uint8(arr[2] * 255), uint8(arr[3] * 255)}
}
