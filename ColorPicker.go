package main

import (
	"fmt"
	"github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"image"
	"image/color"
	"sort"
)

type Gradient struct {
	ticks []GradientTick //This must be insured to be in order of position
}
type GradientTick struct {
	pos   float64
	color color.RGBA
}

func (g *Gradient) Init() {
	//When return see if this is failing to initialize array and store it back or if its the display that doesnt work
	//Sets two values black at 0, white at 1
	g.AddTick(0.0, color.RGBA{0, 0, 0, 0xff})
	g.AddTick(0.5, color.RGBA{0, 0, 255, 0xff})
	g.AddTick(1.0, color.RGBA{255, 255, 255, 0xff})
}

//Get the color at a position
//Stolen straight from the godot game engine
func (g *Gradient) GetColorAt(pos float64) color.RGBA {
	var c color.RGBA
	
	//Make sure pos is in the bounds of the array
	//A binary search would be faster rather than starting at the top
	low := 0
	high := len(g.ticks) - 1
	middle := 0
	for low <= high {
		middle = (low + high) / 2
		var tick GradientTick = g.ticks[middle]
		if tick.pos > pos {
			high = middle - 1 //search low end of array
		} else if tick.pos < pos {
			low = middle + 1 //search high end of array
		} else {
			return tick.color
		}
	}
	
			//return interpolated value
		if (g.ticks[middle].pos > pos) {
			middle--;
		}
		first := middle;
		second := middle + 1;
		
		if (second >= len(g.ticks)) {
			return g.ticks[len(g.ticks) - 1].color;
		}
		if (first < 0) {
			return g.ticks[0].color;
		}
		pointFirst := g.ticks[first];
		pointSecond := g.ticks[second];

		return lerpColor(pointSecond.color, pointFirst.color, (pos- pointFirst.pos) / (pointSecond.pos- pointFirst.pos))//pointFirst.color.lerp(pointSecond.color, (p_offset - pointFirst.offset) / (pointSecond.offset - pointFirst.offset));
	
	
	return c
}
func clamp(x, h, l float64) float64 {
	if x > h {
		x = h
	} else if x < l {
		x = l
	}
	return x
}

//amt is 0-1
func lerpColor(cola, colb color.RGBA, amt float64) color.RGBA {
	r := lerp(cola.R, colb.R, amt)
	g := lerp(cola.G, colb.G, amt)
	b := lerp(cola.B, colb.B, amt)
	a := lerp(cola.A, colb.A, amt)
	return color.RGBA{r, g, b, a}
}

//Sets the color of the tick at the index specified
func (g *Gradient) SetColor(index int, newColor color.RGBA) {
	if index < len(g.ticks) {
		g.ticks[index].color = newColor
	}
}

func (g *Gradient) AddTick(pos float64, color color.RGBA) {
	g.ticks = append(g.ticks, GradientTick{pos, color})
	g.sort()
}

func (g *Gradient) makePreview() *image.RGBA {
	fmt.Println(",aking preview")
	previewSize := 20

	pixels := []uint8{}

	for i := 0; i <= previewSize; i++ {
		amt := float64(i) / float64(previewSize)
		col := g.GetColorAt(amt)
		r, g, b, a := col.RGBA()
		pixels = append(pixels, []uint8{uint8(r), uint8(g), uint8(b), uint8(a)}...)
	}
	fmt.Println("pixels: ", pixels)
	img := image.RGBA{pixels, 1, image.Rectangle{image.Point{0, 0}, image.Point{len(pixels) / 4, 1}}}

	return &img
}

//Sort the gradient by position
func (g *Gradient) sort() {
	sort.SliceStable(g.ticks, func(i, j int) bool { return g.ticks[i].pos < g.ticks[j].pos })
}

//Widget for preview of the widget
type GradientEditorWidget struct {
	previewTexID imgui.TextureID
	grad         *Gradient
}

func GradientEditor(id string, gradient *Gradient, flags int) GradientEditorWidget {

	//img:=gradient.makePreview()
	g := GradientEditorWidget{}
	g.grad = gradient
	g.UpdateTex()
	return g
}

func (gr GradientEditorWidget) Build() {
	imgui.Text("Gradient")
	imgui.ImageButton(gr.previewTexID, imgui.Vec2{100, 10})
}
func (gr *GradientEditorWidget) UpdateTex() {
	renderer := giu.Context.GetRenderer()
	img := gr.grad.makePreview()
	//Discard old texture
	renderer.ReleaseImage(gr.previewTexID)

	texid, err := renderer.LoadImage(img)
	gr.previewTexID = texid
	fmt.Println("ERR: ", err)

}

var pickers []ColorPickerWidget = []ColorPickerWidget{}

func BuildColorPickers() {
	for _, p := range pickers {
		p.Build()
	}
}

func ColorPicker(id string, cols *[3]float32, flags int) ColorPickerWidget {
	p := ColorPickerWidget{id, cols, flags}
	pickers = append(pickers, p)
	fmt.Println("pickers: ", pickers)
	return p
}

type ColorPickerWidget struct {
	id    string
	cols  *[3]float32
	flags int
}

func (c ColorPickerWidget) GetColor() color.RGBA {
	return colFromArr(*(c.cols))
}
func (c ColorPickerWidget) Build() {
	imgui.ColorPicker3(c.id, c.cols, c.flags)
}

func colFromArr(arr [3]float32) color.RGBA {
	return color.RGBA{uint8(arr[0] * 255), uint8(arr[1] * 255), uint8(arr[2] * 255), 0xff}
}

func makeColors() []color.RGBA { //,err
	cols := []color.RGBA{}
	for _, p := range pickers {
		cols = append(cols, p.GetColor())
	}
	return cols //,err if there are no colors
}
