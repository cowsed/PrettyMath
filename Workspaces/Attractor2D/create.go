package attractor2d

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
)

//Helper function that is probably ineffecient but makes the code a bit easier to read/write
func inMap(key [2]int, m map[[2]int]int) bool {
	_, in := m[key]
	return in
}

//CreateImage remakes the plotted points and the image
//In the future separate the plotting of points into the map and rendering that to an image so that color editing can be done faster
func (ws *Workspace) CreateImage() {

	//Create the point map (not a map but ok)
	//var pointMap = [][]int{}
	mapKeys := [][2]int{}
	pointMap := make(map[[2]int]int) // make([][]int, height)

	var maxPoints = 1
	var drawnPoints = 0
	//Generate the attractor

	for i := 0; i < int(ws.numPoints); i++ {
		newx := ws.XExp.BecomeNumber(ws.Variables)
		newy := ws.YExp.BecomeNumber(ws.Variables)

		ws.Variables["x"] = newx
		ws.Variables["y"] = newy

		disx := int(newx*float64(ws.imageWidth)*float64(ws.scaleFactor) + float64(ws.offx))
		disy := int(newy*float64(ws.imageHeight)*float64(ws.scaleFactor) + float64(ws.offy))

		if disx < int(ws.imageWidth) && disy < int(ws.imageHeight) && disx >= 0 && disy >= 0 {
			if inMap([2]int{disx, disy}, pointMap) {
				pointMap[[2]int{disx, disy}] = pointMap[[2]int{disx, disy}] + 1
			} else {
				pointMap[[2]int{disx, disy}] = 1
				mapKeys = append(mapKeys, [2]int{disx, disy})
			}
			drawnPoints++
			maxPoints = maxI(maxPoints, pointMap[[2]int{disx, disy}])
		}

	}
	print("Max Points: ")
	println(maxPoints)
	print("Points Drawn: ")
	println(drawnPoints)
	print("Individual Points: ")
	fmt.Println(len(mapKeys))
	print("Possible Points: ")
	fmt.Println(ws.imageWidth * ws.imageHeight)

	upLeft := image.Point{0, 0}
	lowRight := image.Point{int(ws.imageWidth), int(ws.imageHeight)}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	for x := 0; x < int(ws.imageWidth); x++ {
		for y := 0; y < int(ws.imageHeight); y++ {
			var amt float64 = 0
			if inMap([2]int{x, y}, pointMap) {
				amt = float64(pointMap[[2]int{x, y}]) / float64(maxPoints)
			} else {
				amt = 0
			}
			amt = math.Pow(amt, 1.0/float64(ws.nthRoot))
			newCol := ws.gradient.GetColorAt(amt)
			img.Set(x, y, newCol)
		}
	}
	f, _ := os.Create("out.png")
	png.Encode(f, img)
}

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
