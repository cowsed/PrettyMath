package main

import (
	_ "fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	
	ep "./ExpressionParser"
)

var numPoints int32 = 1_000_000 //_000

var height = 600
var width = 800

var offx = width/2 + 100
var offy = height/2 - 150

var sf = .1

var paramA float64
var paramB float64
var paramC float64
var paramD float64

var colors []color.RGBA

func CreateImage() {

	//Create the point map (not a map but ok)
	var pointMap = [][]int{}
	pointMap = make([][]int, height)

	for i := 0; i < len(pointMap); i++ {
		pointMap[i] = make([]int, width)
	}

	var maxPoints = 1
	var drawnPoints = 0
	//Generate the attractor

	for i := 0; i < int(numPoints); i++ {
		newx := XExp.BecomeNumber()
		newy := YExp.BecomeNumber()

ep.Vars["x"] = newx
		ep.Vars["y"] = newy

		disx := int(newx*float64(width)*sf + float64(offx))
		disy := int(newy*float64(height)*sf + float64(offy))

		if disx < width && disy < height && disx >= 0 && disy >= 0 {
			pointMap[disy][disx] = pointMap[disy][disx] + 1
			drawnPoints++
			maxPoints = maxI(maxPoints, pointMap[disy][disx])
		}

	}
	print("Max Points: ")
	println(maxPoints)
	print("Points Drawn: ")
	println(drawnPoints)
	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {

			amt := float64(pointMap[y][x]) / float64(maxPoints)
			amt = math.Pow(amt, 1.0/2)
			newCol := colGradient.GetColorAt(amt)//lerpColors(amt)
			img.Set(x, y, newCol)
		}
	}
	f, _ := os.Create("out.png")
	png.Encode(f, img)
}


func f4(x, y float64) (float64, float64) {

	a := paramA
	b := paramB
	c := paramC
	d := paramD

	xnew := math.Cos(y*b) + c*math.Cos(x*b)
	ynew := math.Cos(x*a) + d*math.Cos(y*a)
	return xnew, ynew
}



func min(a, b uint8) uint8 {
	if a < b {
		return a
	} else {
		return b
	}
}
func max(a, b uint8) uint8 {
	if a < b {
		return a
	} else {
		return b
	}
}
func minI(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
func maxI(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
func minF(a, b float64) float64 {
	if a < b {
		return a
	} else {
		return b
	}
}
func maxF(a, b float64) float64 {
	if a > b {
		return a
	} else {
		return b
	}
}

func lerp(a, b uint8, amt float64) uint8 {
	return uint8(float64(a)*(amt) + float64(b)*(1-amt))
}
func mult(a, b uint8) uint8 {
	return uint8((float64(a) / 255) * (float64(b) / 255) * 255)
}

func add(a, b uint8) uint8 {
	return min(uint8((float64(a)*.2 + float64(b))), 255)
}

func Add(cola, colb color.RGBA) color.RGBA {

	r := add(cola.R, colb.R)
	g := add(cola.G, colb.G)
	b := add(cola.B, colb.B)
	a := add(cola.A, colb.A)

	return color.RGBA{r, g, b, a}
}
