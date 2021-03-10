package main

import (
	_ "fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
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
		vars["x"] = newx
		newy := YExp.BecomeNumber()
		vars["y"] = newy

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
			newCol := lerpColors(amt)
			img.Set(x, y, newCol)
		}
	}
	f, _ := os.Create("out.png")
	png.Encode(f, img)
}

func lerpColors(amt float64) color.RGBA {
	colors := [2]color.RGBA{}
	//{60, 12, 48, 0xff}//
	//{235, 143, 34, 0xff}//
	colors[0] = color.RGBA{219, 58, 52, 0xff} //{210, 220, 222, 0xff}
	colors[1] = color.RGBA{50, 48, 49, 0xff}  //{60,110,95,0xff}
	//colors[2] = color.RGBA{140,30,45,0xff}
	//colors[3]=color.RGBA{80,50,65,0xff}
	//if (amt>1){println("panic")}
	amtI := minI(int(float64(len(colors))*amt), len(colors)-2)
	c := genCol(colors[amtI], colors[amtI+1], (float64(len(colors))*amt)-float64(amtI))
	return c
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

func genCol(cola, colb color.RGBA, amt float64) color.RGBA {
	amt = maxF(minF(amt, 1.0), 0.0)

	r := lerp(cola.R, colb.R, amt)
	g := lerp(cola.G, colb.G, amt)
	b := lerp(cola.B, colb.B, amt)
	a := lerp(cola.A, colb.A, amt)
	return color.RGBA{r, g, b, a}
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
