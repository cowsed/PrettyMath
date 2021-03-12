package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"./Tools"
	ep "./ExpressionParser"
)

var localGradientEditor tools.GradientEditorWidget
var colGradient tools.Gradient

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

func inMap(key [2]int, m map[[2]int]int) bool {
	_, in := m[key]
	return in
}
func CreateImage() {

	//Create the point map (not a map but ok)
	//var pointMap = [][]int{}
	mapKeys:=[][2]int{}
	pointMap :=make(map[[2]int]int)// make([][]int, height)
	

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
			if inMap([2]int{disx,disy}, pointMap){
				pointMap[[2]int{disx,disy}] = pointMap[[2]int{disx,disy}]+ 1
			} else {
				pointMap[[2]int{disx,disy}] = 1
				mapKeys=append(mapKeys, [2]int{disx,disy})
			}
			drawnPoints++
			maxPoints = maxI(maxPoints, pointMap[[2]int{disx,disy}])
		}

	}
	print("Max Points: ")
	println(maxPoints)
	print("Points Drawn: ")
	println(drawnPoints)
	print("Individual Points: ")
	fmt.Println(len(mapKeys))
	print("Possible Points: ")
	fmt.Println(width*height)
	
	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})
	fmt.Println("Gradient: ", colGradient)
	fmt.Println("GradienttEST: ", colGradient.GetColorAt(0))

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			var amt float64 =0
			if inMap([2]int{x,y}, pointMap){
				amt = float64(pointMap[[2]int{x,y}]) / float64(maxPoints)			
			} else {
				amt=0
			}
			amt = math.Pow(amt, 1.0/float64(nthRoot))
			newCol := colGradient.GetColorAt(amt) 
			
			img.Set(x, y, newCol)
		}
	}
	f, _ := os.Create("out.png")
	png.Encode(f, img)
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
