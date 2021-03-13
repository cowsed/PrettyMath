package tools

import (
	"fmt"
	"github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"image"
	"image/color"
	"sort"
)

//A Gradient that holds color ticks at specific places (The idea and implementation was pretty much stolen straight from the Godot game engine
type Gradient struct {
	ticks []gradientTick //This must be insured to be in order of position
}

//The tick for a gradient. has a position and a color
type gradientTick struct {
	pos   float64
	color color.RGBA
}

//Init creates a default Gradient with default ticks to avoid a slice length of 0
func (g *Gradient) Init() {
	//Sets two values black at 0, white at 1
	g.AddTick(0.0, color.RGBA{0, 0, 0, 0xff})
	g.AddTick(0.3, color.RGBA{255, 0, 255, 0xff})
	g.AddTick(0.6, color.RGBA{255, 255, 255, 0xff})
}

//GetColorAt gets the color at a position and interpolates if there is not a tick there
func (g *Gradient) GetColorAt(pos float64) color.RGBA {
	var c color.RGBA

	//Make sure pos is in the bounds of the array
	//A binary search would be faster rather than starting at the top
	low := 0
	high := len(g.ticks) - 1
	middle := 0
	for low <= high {
		middle = (low + high) / 2
		var tick gradientTick = g.ticks[middle]
		if tick.pos > pos {
			high = middle - 1 //search low end of array
		} else if tick.pos < pos {
			low = middle + 1 //search high end of array
		} else {
			return tick.color
		}
	}

	//return interpolated value
	if g.ticks[middle].pos > pos {
		middle--
	}
	first := middle
	second := middle + 1

	if second >= len(g.ticks) {
		return g.ticks[len(g.ticks)-1].color
	}
	if first < 0 {
		return g.ticks[0].color
	}
	pointFirst := g.ticks[first]
	pointSecond := g.ticks[second]

	return lerpColor(pointSecond.color, pointFirst.color, (pos-pointFirst.pos)/(pointSecond.pos-pointFirst.pos))

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

func lerp(a, b uint8, amt float64) uint8 {
	return uint8(float64(a)*(amt) + float64(b)*(1-amt))
}

//SetColor sets the color of the tick at the index specified
func (g *Gradient) SetColor(index int, newColor color.RGBA) {
	if index < len(g.ticks) {
		g.ticks[index].color = newColor
	}
}

//AddTick adds a tick with a position and a color
func (g *Gradient) AddTick(pos float64, color color.RGBA) {
	g.ticks = append(g.ticks, gradientTick{pos, color})
	g.sort()
}

//makePreview creates a 1 by previewSize image for previewing the gradient in a widget
func (g *Gradient) makePreview() *image.RGBA {
	//Horizontal resolution of the preview
	previewSize := 20
	pixels := []uint8{}
	for i := 0; i <= previewSize; i++ {
		amt := float64(i) / float64(previewSize)
		col := g.GetColorAt(amt)
		r, g, b, a := col.RGBA()
		pixels = append(pixels, []uint8{uint8(r), uint8(g), uint8(b), uint8(a)}...)
	}
	img := image.RGBA{pixels, 1, image.Rectangle{image.Point{0, 0}, image.Point{len(pixels) / 4, 1}}}
	return &img
}

//Sort the gradient by position. Necessary for correcct interpolation
func (g *Gradient) sort() {
	sort.SliceStable(g.ticks, func(i, j int) bool { return g.ticks[i].pos < g.ticks[j].pos })
}

//GradientEditorWidget is a widget for viewing and editing a gradient
type GradientEditorWidget struct {
	previewTexID imgui.TextureID
	grad         *Gradient
}

//GradientEditor creates a GradientEditor for use in immediate mode gui
func GradientEditor(id string, gradient *Gradient, flags int) GradientEditorWidget {

	//img:=gradient.makePreview()
	g := GradientEditorWidget{}
	g.grad = gradient
	g.UpdateTex()
	return g
}

//Build provides the build function for immediate mode
func (gr GradientEditorWidget) Build() {

	var availableWidth float32 = 100.0 //TODO make this take up the correct amount of space
	imgui.Text("Gradient")
	p := giu.GetCursorScreenPos()

	imgui.ImageButton(gr.previewTexID, imgui.Vec2{availableWidth, 10})
	drawList := imgui.GetWindowDrawList()

	tickHeight := 14
	tickBgCol := giu.ToVec4Color(color.RGBA{uint8(28), uint8(36), uint8(43), uint8(0xff)})
	//imgui.Vec4{1,0,0,1}//
	tickBorder := 2
	topOffset := 10
	tickSize := imgui.Vec2{float32(tickBorder*2 + topOffset), float32(tickHeight + topOffset + tickBorder*2)}
	for i, t := range gr.grad.ticks {
		col := giu.ToVec4Color(t.color)
		pos := float32(t.pos) * availableWidth

		pmin := p.Add(image.Point{int(pos) - tickBorder, 10 - tickBorder})
		pmax := p.Add(image.Point{10 + int(pos) + tickBorder, tickHeight + 10 + tickBorder})
		drawList.AddRectFilled(giu.ToVec2(pmin), giu.ToVec2(pmax), tickBgCol, 0, 5)

		pmin2 := p.Add(image.Point{int(pos), 10})
		pmax2 := p.Add(image.Point{10 + int(pos), tickHeight + 10})
		drawList.AddRectFilled(giu.ToVec2(pmin2), giu.ToVec2(pmax2), col, 0, 5)

		imgui.SetCursorScreenPos(giu.ToVec2(p.Add(image.Point{int(availableWidth * float32(t.pos)), 0})))
		if imgui.ButtonV("ButtonX"+string(i), tickSize) {
			fmt.Println("clicked: ", i)
		}
		if imgui.IsItemHovered() {
			fmt.Println("Hovered")
		}
	}
}

//UpdateTex gets a new image from the gradient and deals with registering it as a textyre
func (gr *GradientEditorWidget) UpdateTex() {
	renderer := giu.Context.GetRenderer()
	img := gr.grad.makePreview()
	//Discard old texture
	renderer.ReleaseImage(gr.previewTexID)

	texid, err := renderer.LoadImage(img)
	gr.previewTexID = texid
	fmt.Println("Texture Update Error ERR: ", err)

}
