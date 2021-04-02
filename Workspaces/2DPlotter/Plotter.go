package Plotter

import (
	"fmt"
	"github.com/AllenDang/giu"

	ep "github.com/cowsed/PrettyMath/ExpressionParser"
)

type Workspace struct {
	amOpen  bool
	onClose func()

	equations []string
	compiledTo []string
	xs [][]float64
	ys [][]float64
}

var xs = []float64{1, 2, 3, 4}
var ys = []float64{2, 1, 2, 1}

//Init initializes a new plotter workspace
func Init(onCloseFunc func()) Workspace {
	ws := Workspace{
		true,
		onCloseFunc,
		[]string{"sin(x)", "x^2"},
		[]string{"",""},
		[][]float64{},
		[][]float64{},
	}

	ws.updateEquations()
	return ws
}
func (ws *Workspace) updateEquations(){
	ws.xs=nil
	ws.ys=nil
	ws.xs=make([][]float64,len(ws.equations))
	ws.ys=make([][]float64,len(ws.equations))

	for i:=0; i<len(ws.equations); i++{
		xs,ys,s:=makeEquation(ws.equations[i],0,100)
		ws.xs[i]=xs
		ws.ys[i]=ys
		ws.compiledTo[i]=s
	}
}

func (ws *Workspace) makeEquationEditor(i int, v interface{}) giu.Widget {
	item := giu.Layout{
		giu.InputText("y= # "+string(i), &ws.equations[i]).OnChange(ws.updateEquations), //.OnChange(ws.UpdateImageAuto),
	}
	return item
}
func makeEquation(exp string, min, max float64) ([]float64, []float64, string) {
	//These will be calculated in the future automatically based on zoom level and such
	numPoints := 100
	step := 0.1

	vars := map[string]float64{"x": 0}
	ee := ep.ParseExpression(exp, vars)

	xs := make([]float64, numPoints)
	ys := make([]float64, numPoints)

	x := min
	for i := 0; i < numPoints; i++ {
		xs[i] = x
		vars["x"] = x
		y := ee.BecomeNumber(vars)
		ys[i] = y

		x += step
	}
	return xs, ys, ee.BecomeString()
}

func (ws *Workspace) makeEquationPlots(i int) giu.PlotWidget{
	//Figure out why these only show y=x
	fmt.Println(ws.ys[i][0:5])
	plot := giu.PlotLineXY(ws.equations[i]+" "+string(i), ws.xs[i], ws.ys[i])
	return plot
}

func (ws *Workspace) Build() {
	//Close when necessary
	if !ws.amOpen {
		println("Closing\n\n\n\n\n\n\n")
		ws.onClose()
	}

	//Needed for rangebuilder (should probably find a way around this, it probably isnt very memory effecient
	eqInterface := make([]interface{}, len(ws.equations))
	//graphsInterface := make([]interface{}, len(ws.equations))


	//Make plots
	plots:=make([]giu.PlotWidget, len(ws.equations))
	for i,_ := range(plots){
		plots[i]=ws.makeEquationPlots(i)
	}
	//Equation Editor
	editorSide := giu.Group().Layout(
		giu.Label("Editor Side"),
		giu.RangeBuilder("Buttons in one line", eqInterface, ws.makeEquationEditor),
	)
	//Plots
	plotSide := giu.Group().Layout(
		//giu.RangeBuilder("Plots", graphsInterface, ws.makeEquationPlots),
		giu.Plot("Plots").Plots(
			plots...
		).Size(-1,-1),
	)
	//All of it
	giu.TabItem("Plotter").Layout(
		giu.SplitLayout("MainSplit", giu.DirectionHorizontal, true, 300,
			editorSide,
			plotSide,
		),
	).IsOpen(&ws.amOpen).Build()
}
