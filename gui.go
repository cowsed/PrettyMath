package main

import (
	"fmt"
	g "github.com/AllenDang/giu"

	"image"
	//"image/color"
	"strconv"
	//"errors"

	ep "./ExpressionParser"
)

var (
	texture *g.Texture
)

var localGradientEditor GradientEditorWidget
var colGradient Gradient

var NewWindowOpen = true
var Attractor2DOpen = true

var connectPoints = false

var autoUpdate bool = false

var testCol [3]float32
var testCol2 [3]float32
var testCol3 [3]float32
var testCol4 [3]float32

var aString string = "0.65343"
var bString string = "0.7345345"
var cString string = "1.3"
var dString string = "1.4"

var x0 float64 = 0.1
var y0 float64 = 0.1
var x0Str string = "0.1"
var y0Str string = "0.1"

var XExpStr string = "sin(x*y/b)*y+cos(a*x-y)"
var YExpStr string = "x+sin(y)/b"

var XExp ep.EquationElement
var YExp ep.EquationElement

var XExpRep string = "--Regenerate to show compiled equation--"
var YExpRep string = "--Regenerate to show compiled equation--"

var imageDisplayX int32 = 800
var imageDisplayY int32 = 600

var offxPer float32 = .5
var offyPer float32 = .5

var scaleFactor float32 = .1

func updateParams() {
	_, err := strconv.ParseFloat("a", 64)
	paramA, err = strconv.ParseFloat(aString, 64)
	paramB, err = strconv.ParseFloat(bString, 64)
	paramC, err = strconv.ParseFloat(cString, 64)
	paramD, err = strconv.ParseFloat(dString, 64)
	fmt.Println("Param ERR:", err)

	x0, err = strconv.ParseFloat(x0Str, 64)
	y0, err = strconv.ParseFloat(y0Str, 64)
	//Reset Vars
	ep.Vars = nil
	ep.Vars = make(map[string]float64)
	ep.Vars["a"] = paramA
	ep.Vars["b"] = paramB
	ep.Vars["c"] = paramC
	ep.Vars["d"] = paramD

	ep.Vars["x"] = float64(x0)
	ep.Vars["y"] = float64(y0)

	offx = int(float32(imageDisplayX) * (offxPer))
	offy = int(float32(imageDisplayY) * (offyPer))

	width = int(imageDisplayX)
	height = int(imageDisplayY)
	sf = float64(scaleFactor)

	//Compile Expressions
	XExp = ep.ParseExpression(XExpStr)
	YExp = ep.ParseExpression(YExpStr)

	//Representations of The Expressions

	XExpRep = XExp.BecomeString()
	YExpRep = YExp.BecomeString()

	colors = makeColors()
}

func UpdateImage() {
	if autoUpdate {
		CreateLoadImage()
	}
}
func CreateLoadImage() {
	updateParams()
	CreateImage()
	loadImage()
	g.Update()
}

func ExpandAll() {

}

func loop() {
	localGradientEditor=GradientEditor("Gradient", &colGradient, 0)
	localGradientEditor.UpdateTex()
	
	//Reset list of color Pickers (probably not the best way to do this)
	pickers = nil
	//Ensure the new window dialog is open
	NewWindowOpen = true
	fullcanvas := g.Layout{
		g.Custom(func() {
			canvas := g.GetCanvas()
			pos := g.GetCursorScreenPos()
			//color := color.RGBA{200, 75, 75, 255}
			//canvas.AddLine(pos, pos.Add(image.Pt(100, 100)), color, 1)

			if texture != nil {
				canvas.AddImage(texture, pos, pos.Add(image.Pt(int(imageDisplayX), int(imageDisplayY))))
			}
		}),
	}
	parameterInput := g.Group().Layout(

		g.Separator(),
		g.TreeNode("Parameters").Layout(
			g.InputText("A", &aString).OnChange(UpdateImage), g.Tooltip("Parameter a"),
			g.InputText("B", &bString).OnChange(UpdateImage), g.Tooltip("Parameter b"),
			g.InputText("C", &cString).OnChange(UpdateImage), g.Tooltip("Parameter c"),
			g.InputText("D", &dString).OnChange(UpdateImage), g.Tooltip("Parameter d"),
			g.Separator(),
			//Initial x,y s
			g.InputText("X0", &x0Str), g.Tooltip("Initial X Value"),
			g.InputText("Y0", &y0Str), g.Tooltip("Initial Y Value"),
			g.Checkbox("Connect Points", &connectPoints), g.Tooltip("If true, connects all the points with lines(but not now bc i havent gotten there yet)"),
			g.Separator(),
		).Flags(g.TreeNodeFlagsFramed), g.Tooltip("The paramaters for the equations"),
		g.Separator(),

		g.TreeNode("Equations").Layout(
			g.InputText("XExp", &XExpStr), g.Tooltip("The expression for newx="),
			g.InputText("YExp", &YExpStr), g.Tooltip("The expression for newy="),
			g.Separator(),

			g.TreeNode("Compiled To").Layout(
				g.Label(XExpRep), g.Tooltip("What the computer sees for XExp"),
				g.Label(YExpRep), g.Tooltip("What the computer sees for YExp"),
			).Flags(g.TreeNodeFlagsFramed), g.Tooltip("What the computer sees"),
		).Flags(g.TreeNodeFlagsFramed), g.Tooltip("The equations that make the image"),
		g.Separator(),

		g.TreeNode("Display Parameters").Layout(
			g.DragInt("Image X", &imageDisplayX, 0, 2000), g.Tooltip("The width of the image \n(before regeneration just scales the previous image.\nwhen it regenerates it will create an image of this size)"),
			g.DragInt("Image Y", &imageDisplayY, 0, 2000), g.Tooltip("The height of the image \n(before regeneration just scales the previous image.\nwhen it regenerates it will create an image of this size)"),
			g.InputFloat("Offset X%", &offxPer),        //.OnChange(UpdateImage),
			g.InputFloat("Offset Y%", &offyPer),        //.OnChange(UpdateImage),
			g.InputFloat("Scale Factor", &scaleFactor), //.OnChange(UpdateImage),
			g.DragInt("Num Points", &numPoints, 0, 1_000_000_000),
			g.Separator(),
		).Flags(g.TreeNodeFlagsFramed), g.Tooltip("Parameters for the output of the image"),
		g.Separator(),
		//g.CustomWidget{imgui.ColorPicker3("Color1Picker",&testCol, 0)}),
		g.TreeNode("Colors").Layout(
			ColorPicker("First", &testCol, 0),
			//ColorPicker("Second", &testCol2, 0),
			//ColorPicker("Third", &testCol3, 0),
			//ColorPicker("Fourth", &testCol4, 0),
		),
		localGradientEditor,
	)

	g.SingleWindow("Images").Layout(
		g.TabBar("TabBar").Layout(
			g.TabItem("2D Attractors").Layout(
				g.Line(
					g.Button("Regenerate").OnClick(CreateLoadImage), g.Tooltip("Regenerate Image"),
					g.Checkbox("Auto-update", &autoUpdate), g.Tooltip("Update On Parameter Change"),
					g.Button("Expand All").OnClick(ExpandAll), g.Tooltip("Expand all parameter windows"),
				),
				g.SplitLayout("MainSplit", g.DirectionHorizontal, true, 300,
					g.Layout{

						parameterInput,
					},
					fullcanvas,
				),
			).IsOpen(&Attractor2DOpen),
			g.TabItem("+").Layout(
				g.Label("Shouldnt have closed that other window, huh bud"),
			),
		),
	)
	fmt.Println(colGradient)
}

func main() {
	colGradient.Init()

	wnd := g.NewMasterWindow("Fun Graphics Stuff", 1200, 800, 0, nil)

	loadImage()

	wnd.Run(loop)

}
func loadImage() {
	img, _ := g.LoadImage("out.png")
	go func() {
		texture, _ = g.NewTextureFromRgba(img)
	}()
}
