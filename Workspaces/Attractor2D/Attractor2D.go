package attractor2d

import (
	"fmt"
	"image"
	"strconv"

	ep "../../ExpressionParser"
	"../../Tools"
	g "github.com/AllenDang/giu"
)

//Workspace is the workspace for generating 2-dimensional Attracors
type Workspace struct {
	//General Settings
	amOpen        bool //should be true usually
	connectPoints bool //= false
	autoUpdate    bool //= false
	numPoints     int32
	onClose       func()
	//Parameters
	//Editor Versions
	aString string //= "0.65343"
	bString string //= "0.7345345"
	cString string //= "1.3"
	dString string //= "1.4"
	//Value Versions
	paramA float64
	paramB float64
	paramC float64
	paramD float64

	//Initial Points
	x0    float64 //= 0.1
	y0    float64 // = 0.1
	x0Str string  //= "0.1"
	y0Str string  //= "0.1"

	//Scale and translations
	scaleFactor float32
	offxPer     float32
	offyPer     float32
	offx        int
	offy        int

	//Expression Stuff
	XExpStr string //= "sin(x*y/b)*y+cos(a*x-y)"
	YExpStr string // = "x+sin(y)/b"
	XExp    ep.EquationElement
	YExp    ep.EquationElement
	XExpRep string //= "--Regenerate to show compiled equation--"
	YExpRep string // = "--Regenerate to show compiled equation--"

	Variables map[string]float64

	//Colors
	gradient tools.Gradient
	nthRoot  float32 //= 2

	//Output Info
	imageWidth  int32 // 800
	imageHeight int32 //= 600
	image       image.RGBA
	imageTex    *g.Texture

	displayScale float32 // =0.75

}

//Init creates a new 2d attractor workspace with default parameters
func Init(onCloseFunc func()) Workspace {
	gradient := tools.Gradient{}
	gradient.Init()

	return Workspace{
		amOpen:        true,
		connectPoints: false,
		autoUpdate:    false,
		onClose:       onCloseFunc,
		numPoints:     1_000_000,
		aString:       "0.65343",
		bString:       "0.7345345",
		cString:       "1.3",
		dString:       "1.4",
		gradient:      gradient,

		x0Str:       "0.1",
		y0Str:       "0.1",
		scaleFactor: 0.1,
		offxPer:     0.5,
		offyPer:     0.5,

		XExpStr: "sin(x*y/b)*y+cos(a*x-y)",
		YExpStr: "x+(sin(y)/b)",

		XExpRep: "--Regenerate to show compiled equation--",
		YExpRep: "--Regenerate to show compiled equation--",

		nthRoot:      2,
		imageWidth:   1366,
		imageHeight:  768,
		displayScale: 0.75,
	}
}

//Takes all the editors and their values and updates the values to be sent to the rendering function
func (ws *Workspace) updateParams() {
	_, err := strconv.ParseFloat("a", 64)
	ws.paramA, err = strconv.ParseFloat(ws.aString, 64)
	ws.paramB, err = strconv.ParseFloat(ws.bString, 64)
	ws.paramC, err = strconv.ParseFloat(ws.cString, 64)
	ws.paramD, err = strconv.ParseFloat(ws.dString, 64)
	fmt.Println("Param ERR:", err)

	ws.x0, err = strconv.ParseFloat(ws.x0Str, 64)
	ws.y0, err = strconv.ParseFloat(ws.y0Str, 64)
	//Reset Vars
	ws.Variables = nil
	ws.Variables = make(map[string]float64)
	ws.Variables["a"] = ws.paramA
	ws.Variables["b"] = ws.paramB
	ws.Variables["c"] = ws.paramC
	ws.Variables["d"] = ws.paramD

	ws.Variables["x"] = ws.x0
	ws.Variables["y"] = ws.y0

	ws.offx = int(float32(ws.imageWidth) * (ws.offxPer))
	ws.offy = int(float32(ws.imageHeight) * (ws.offyPer))

	//Compile Expressions
	ws.XExp = ep.ParseExpression(ws.XExpStr, ws.Variables)
	ws.YExp = ep.ParseExpression(ws.YExpStr, ws.Variables)

	//Representations of The Expressions

	ws.XExpRep = ws.XExp.BecomeString()
	ws.YExpRep = ws.YExp.BecomeString()

}

//UpdateImageAuto updates the rendered image if auto update is on
func (ws *Workspace) UpdateImageAuto() {
	if ws.autoUpdate {
		ws.CreateLoadImage()
	}
}

//CreateLoadImage completely rerenders and reloads the render
func (ws *Workspace) CreateLoadImage() {
	ws.updateParams()
	ws.CreateImage()
	ws.loadImage()
	g.Update()
}

//Loads the image from an image.RGBA (for now a file) into a texture to display
func (ws *Workspace) loadImage() {
	img, _ := g.LoadImage("out.png")
	go func() {
		ws.imageTex, _ = g.NewTextureFromRgba(img)
	}()
}

//Build builds the workspace for use with giu
func (ws *Workspace) Build() {
	if !ws.amOpen {
		fmt.Println("Closing\n\n\n\n\n\n\n")
		ws.onClose()
	}

	fullcanvas := g.Layout{
		g.Custom(func() {
			canvas := g.GetCanvas()
			pos := g.GetCursorScreenPos()
			if ws.imageTex != nil {
				canvas.AddImage(ws.imageTex, pos, pos.Add(image.Pt(int(float32(ws.imageWidth)*ws.displayScale), int(float32(ws.imageHeight)*ws.displayScale))))
			}
		}),
	}
	EditorPanel := g.Group().Layout(
		g.Separator(),
		g.TreeNode("Parameters").Layout(
			g.InputText("A", &ws.aString).OnChange(ws.UpdateImageAuto), g.Tooltip("Parameter a"),
			g.InputText("B", &ws.bString).OnChange(ws.UpdateImageAuto), g.Tooltip("Parameter b"),
			g.InputText("C", &ws.cString).OnChange(ws.UpdateImageAuto), g.Tooltip("Parameter c"),
			g.InputText("D", &ws.dString).OnChange(ws.UpdateImageAuto), g.Tooltip("Parameter d"),
			g.Separator(),
			//Initial x,y s
			g.InputText("X0", &ws.x0Str), g.Tooltip("Initial X Value"),
			g.InputText("Y0", &ws.y0Str), g.Tooltip("Initial Y Value"),
			g.Checkbox("Connect Points", &ws.connectPoints), g.Tooltip("If true, connects all the points with lines(but not now bc i havent gotten there yet)"),
			g.Separator(),
		).Flags(g.TreeNodeFlagsFramed), g.Tooltip("The paramaters for the equations"),
		g.Separator(),

		g.TreeNode("Equations").Layout(
			g.InputText("=newx", &ws.XExpStr), g.Tooltip("The expression for newx="),
			g.InputText("=newy", &ws.YExpStr), g.Tooltip("The expression for newy="),
			g.Separator(),

			g.TreeNode("Compiled To").Layout(
				g.Label(ws.XExpRep), g.Tooltip("What the computer sees for XExp"),
				g.Label(ws.YExpRep), g.Tooltip("What the computer sees for YExp"),
			).Flags(g.TreeNodeFlagsFramed), g.Tooltip("What the computer sees"),
		).Flags(g.TreeNodeFlagsFramed), g.Tooltip("The equations that make the image"),
		g.Separator(),

		g.TreeNode("Image Parameters").Layout(
			g.DragInt("Image Width", &ws.imageHeight, 0, 4000), g.Tooltip("The width of the outputted image"),
			g.DragInt("Image Height", &ws.imageWidth, 0, 4000), g.Tooltip("The height of the image outputted image)"),
			g.InputFloat("Offset X%", &ws.offxPer), g.Tooltip("X translation of output (0 is left, 1 is right"),
			g.InputFloat("Offset Y%", &ws.offyPer), g.Tooltip("Y translation of output (0 is top, 1 is bottom"),
			g.InputFloat("Scale Factor", &ws.scaleFactor), g.Tooltip("Scale factor on the points in the output"),
			g.DragInt("Num Points", &ws.numPoints, 0, 1_000_000_000), g.Tooltip("Number of points in the output image"),
			g.Separator(),
		).Flags(g.TreeNodeFlagsFramed), g.Tooltip("Parameters for the output of the image"),
		g.TreeNode("Display Parameters").Layout(
			g.InputFloat("Display Scale", &ws.displayScale), g.Tooltip("Scale of the output file size to display"),
		).Flags(g.TreeNodeFlagsFramed),
		g.Separator(),
		g.TreeNode("Colors").Layout(
			g.InputFloat("nthRt(points)", &ws.nthRoot), g.Tooltip("n for the nth root of normalized points in a pixel"),
			tools.GradientEditor("Gradient Editor", &ws.gradient, 0),
		).Flags(g.TreeNodeFlagsFramed),
	)

	g.TabItem("2D Attractors").Layout(

		g.Line(
			g.Button("Regenerate").OnClick(ws.CreateLoadImage), g.Tooltip("Regenerate Image"),
			g.Checkbox("Auto-update", &ws.autoUpdate), g.Tooltip("Update On Parameter Change"),
			//g.Button("Expand All").OnClick(ExpandAll), g.Tooltip("Expand all parameter windows"),
		),
		g.SplitLayout("MainSplit", g.DirectionHorizontal, true, 300,
			g.Layout{
				EditorPanel,
			},
			fullcanvas,
		),
	).IsOpen(&ws.amOpen).Build()

}
