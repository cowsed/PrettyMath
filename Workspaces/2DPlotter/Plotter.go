package Plotter

import (
	"fmt"

	"github.com/AllenDang/giu"
	parser "github.com/cowsed/Parser"
)

type Workspace struct {
	amOpen  bool
	onClose func()

	equations     []string
	compiled      []parser.Expression
	equationsMade []bool
	xs            [][]float64
	ys            [][]float64
}

//Init initializes a new plotter workspace
func Init(onCloseFunc func()) Workspace {
	ws := Workspace{
		amOpen:        true,
		onClose:       onCloseFunc,
		equations:     []string{"sin(x)", "cos(x)"},
		compiled:      []parser.Expression{},
		equationsMade: []bool{},
	}
	ws.MakeSlicesToSize()

	return ws
}

func (ws *Workspace) UpdateExpressionAndGraph() {
	for i := 0; i < len(ws.equations); i++ {
		if !ws.equationsMade[i] {
			ws.CompileExpression(i)
			ws.UpdateGraph(i)
		}
	}
}

func (ws *Workspace) CompileExpression(i int) error {
	e, err := parser.ParseExpression(ws.equations[i])
	if err != nil {
		return err
	}
	ws.compiled[i] = e
	return nil
}
func (ws *Workspace) EvaluateExpressionRange(i int, min, max float64) ([]float64, []float64) {

	var numPoints = 400
	step := (max - min) / float64(numPoints)
	xs := make([]float64, numPoints)
	ys := make([]float64, numPoints)
	j := 0
	for x := min; x < max; x += step {
		xs[j] = x
		ys[j] = ws.compiled[i].Evaluate(map[string]float64{"x": x})
		j++
	}
	return xs, ys
}

func (ws *Workspace) UpdateGraph(i int) {
	ws.xs[i] = make([]float64, len(ws.equations))
	ws.ys[i] = make([]float64, len(ws.equations))
	ws.xs[i], ws.ys[i] = ws.EvaluateExpressionRange(i, 0, 100)

}
func (ws *Workspace) makeEquationPlots(i int) giu.PlotWidget {
	if !ws.equationsMade[i] {
		err := ws.CompileExpression(i)
		if err != nil {
			fmt.Println("Compiling Error", err.Error())
			return giu.PlotLine("Errored Plot", []float64{})
		}
		ws.UpdateGraph(i)
	}
	plot := giu.PlotLineXY(fmt.Sprintf("%s 3%d", ws.compiled[i].String(), i), ws.xs[i], ws.ys[i])
	return plot
}

func (ws *Workspace) MakeSlicesToSize() {
	if len(ws.xs) != len(ws.equations) {
		ws.xs = make([][]float64, len(ws.equations))
	}
	if len(ws.ys) != len(ws.equations) {
		ws.ys = make([][]float64, len(ws.equations))
	}
	if len(ws.compiled) != len(ws.equations) {
		ws.compiled = make([]parser.Expression, len(ws.equations))
	}
	if len(ws.equationsMade) != len(ws.equations) {
		ws.equationsMade = make([]bool, len(ws.equations))
	}

}

func (ws *Workspace) makeEquationEditor(i int, v interface{}) giu.Widget {
	item := giu.Layout{
		giu.InputText("y= # "+fmt.Sprint(i), &ws.equations[i]).OnChange(func() {
			ws.equationsMade[i] = false
			ws.UpdateExpressionAndGraph()
		}), //.OnChange(ws.UpdateImageAuto),
	}
	return item
}
func (ws *Workspace) Build() {
	//Close when necessary
	if !ws.amOpen {
		println("Closing\n\n\n\n\n\n\n")
		ws.onClose()
	}
	ws.MakeSlicesToSize()
	//Needed for rangebuilder (should probably find a way around this, it probably isnt very memory effecient
	eqInterface := make([]interface{}, len(ws.equations))
	//graphsInterface := make([]interface{}, len(ws.equations))

	//Make plots
	plots := make([]giu.PlotWidget, len(ws.equations))
	for i := range plots {
		plots[i] = ws.makeEquationPlots(i)
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
			plots...,
		).Size(-1, -1),
	)
	//All of it
	giu.TabItem("Plotter").Layout(
		giu.SplitLayout("MainSplit", giu.DirectionHorizontal, true, 300,
			editorSide,
			plotSide,
		),
	).IsOpen(&ws.amOpen).Build()
}
