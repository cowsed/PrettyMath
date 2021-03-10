package main

import (
"fmt"
	g "github.com/AllenDang/giu"
	"image"
	//"image/color"
	"strconv"
	//"errors"
)

var (
	texture *g.Texture
)

var noteString = "Notes"
var autoUpdate bool = false

var aString string = "1.1"
var bString string = "1.2"
var cString string = "1.3"
var dString string = "1.4"

var x0 float64 = 0.1
var y0 float64 = 0.1
var x0Str string = "0.1"
var y0Str string = "0.1"

var XExpStr string = "sin(y*b)+c*sin(x*b)"
var YExpStr string = "sin(x*a)+d*sin(y*a)"

var XExp EquationElement
var YExp EquationElement

var imageDisplayX int32 =800
var imageDisplayY int32 =600

var offxPer float32 = .5
var offyPer float32 = .5

var scaleFactor float32 = .1
//var numPoints int32 = 1_000_000

func updateParams(){
	_,err:=strconv.ParseFloat("a",64)
	paramA,err=strconv.ParseFloat(aString,64)
	paramB,err=strconv.ParseFloat(bString,64)
	paramC,err=strconv.ParseFloat(cString,64)
	paramD,err=strconv.ParseFloat(dString,64)
	fmt.Println("Param ERR:",err)

	x0,err=strconv.ParseFloat(x0Str,64)
	y0,err=strconv.ParseFloat(y0Str,64)
	//Reset Vars
	vars=nil
	vars=make(map[string]float64)
	vars["a"]=paramA
	vars["b"]=paramB
	vars["c"]=paramC
	vars["d"]=paramD

	vars["x"]=float64(x0)
	vars["y"]=float64(y0)
		
	offx=int(float32(imageDisplayX)*(offxPer))
	offy=int(float32(imageDisplayY)*(offyPer))
	
	width=int(imageDisplayX)
	height=int(imageDisplayY)
	sf=float64(scaleFactor)
	
	//Compile Expressions
	XExp=parseExpression(XExpStr)
	YExp=parseExpression(YExpStr)
	
}
func UpdateImage(){
	if autoUpdate{
		CreateLoadImage()
	}
}
func CreateLoadImage() {
	//go func(){
		updateParams()
		CreateImage()
		loadImage()
		g.Update()
	//}()
}


func loop() {
	fullcanvas := g.Layout{
		g.Custom(func() {
			canvas := g.GetCanvas()
			pos := g.GetCursorScreenPos()
			//color := color.RGBA{200, 75, 75, 255}
			//canvas.AddLine(pos, pos.Add(image.Pt(100, 100)), color, 1)

			if texture != nil {
				canvas.AddImage(texture, pos, pos.Add(image.Pt(int(imageDisplayX),int(imageDisplayY)) ))
			}
		}),
	}
	parameterInput := g.Layout{
		g.Line(
			g.Button("Regenerate").OnClick(CreateLoadImage),
			g.Checkbox("Auto-update",&autoUpdate),
		),
		g.Separator(),
		g.TreeNode("Parameters").Layout(
			g.InputText("A", &aString).OnChange(UpdateImage),
			g.InputText("B", &bString).OnChange(UpdateImage),
			g.InputText("C", &cString).OnChange(UpdateImage),
			g.InputText("D", &dString).OnChange(UpdateImage),
		
			g.Separator(),
		
			g.InputText("X0", &x0Str),
			g.InputText("Y0", &y0Str),
		),
		g.TreeNode("Notes").Layout(
			g.InputTextMultiline("", &noteString),
		),
		g.Separator(),
		g.TreeNode("Equations").Layout(
			g.InputText("XExp", &XExpStr),
			g.InputText("YExp", &YExpStr),
		),
		g.Separator(),

		g.TreeNode("Display").Layout(
			g.DragInt("Image X", &imageDisplayX, 0,2000),
			g.DragInt("Image Y", &imageDisplayY, 0,2000),
			g.InputFloat("Offset X%", &offxPer),//.OnChange(UpdateImage),
			g.InputFloat("Offset Y%", &offyPer),//.OnChange(UpdateImage),
		),

		g.InputFloat("Scale Factor", &scaleFactor),//.OnChange(UpdateImage),
		g.DragInt("Num Points", &numPoints, 0,9_000_000),

	}
	g.SingleWindow("canvas").Layout(
		g.SplitLayout("MainSplit", g.DirectionHorizontal, true, 300,
			g.Layout{
				parameterInput,
			},
			fullcanvas,
		),
	)
}

func main() {
	wnd := g.NewMasterWindow("2D Attractors", 600, 600, 0, nil)

	loadImage()

	wnd.Run(loop)
}
func loadImage() {
	img, _ := g.LoadImage("out.png")
	go func() {
		texture, _ = g.NewTextureFromRgba(img)
	}()
}
