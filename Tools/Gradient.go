package tools

import (
	"fmt"
	"image"
	"image/color"
	"sort"

	"github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
)

//A Gradient that holds color ticks at specific places (The idea and implementation was pretty much stolen straight from the Godot game engine
type Gradient struct {
	ticks []gradientTick //This must be insured to be in order of position
}

//The tick for a gradient. has a position and a color
type gradientTick struct {
	pos   float64
	color [4]float32
}

//GradientInit creates a default Gradient with default ticks to avoid a slice length of 0
func GradientInit() Gradient {
	g := Gradient{}
	//Sets two values black at 0, white at 1
	g.AddTick(0.0, [4]float32{0, 0, 0, 1})
	g.AddTick(0.3, [4]float32{1, 0, 1, 1})
	g.AddTick(0.6, [4]float32{1, 1, 1, 1})
	return g
}

//GetColorAt gets the color at a position and interpolates if there is not a tick there
func (g *Gradient) GetColorAt(pos float64) [4]float32 {
	//var c color.RGBA

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

	//return c
}
func clamp(x, l, h float64) float64 {
	if x > h {
		x = h
	} else if x < l {
		x = l
	}
	return x
}

//amt is 0-1
func lerpColor(cola, colb [4]float32, amt float64) [4]float32 {
	r := lerp(cola[0], colb[0], amt)
	g := lerp(cola[1], colb[1], amt)
	b := lerp(cola[2], colb[2], amt)
	a := lerp(cola[3], colb[3], amt)
	return [4]float32{r, g, b, a}
}

func lerp(a, b float32, amt float64) float32 {
	return (a*float32(amt) + b*float32(1-amt))
}

//SetColor sets the color of the tick at the index specified
func (g *Gradient) SetColor(index int, newColor [4]float32) {
	if index < len(g.ticks) {
		g.ticks[index].color = newColor
	}
}

//SetPos sets the position of a tick specified at a certain index
func (g *Gradient) SetPos(index int, newPos float64) {
	if index < len(g.ticks) {
		g.ticks[index].pos = newPos
	}
	g.sort()
}

//AddTick adds a tick with a position and a color
func (g *Gradient) AddTick(pos float64, color [4]float32) {
	g.ticks = append(g.ticks, gradientTick{pos, color})
	g.sort()
}

func (g *Gradient) RemoveTick(i int) {
	g.ticks[i] = g.ticks[len(g.ticks)-1]     // Copy last element to index i.
	g.ticks[len(g.ticks)-1] = gradientTick{} // Erase last element (write zero value).
	g.ticks = g.ticks[:len(g.ticks)-1]       // Truncate slice.

	//Make it not empty
	if len(g.ticks) == 0 {
		g.AddTick(0.5, [4]float32{0, 0, 0, 1})
	}
}

//Copy deep copies the gradient
func (g *Gradient) Copy() Gradient {
	newg := GradientInit()
	newg.ticks = make([]gradientTick, len(g.ticks))
	copy(newg.ticks, g.ticks)
	return newg
}

//makePreview creates a 1 by previewSize image for previewing the gradient in a widget
func (g *Gradient) makePreview() *image.RGBA {
	//Horizontal resolution of the preview
	previewSize := 20
	pixels := []uint8{}
	for i := 0; i <= previewSize; i++ {
		amt := float64(i) / float64(previewSize)
		col := g.GetColorAt(amt)
		r, g, b, a := col[0], col[1], col[2], col[3]
		pixels = append(pixels, []uint8{uint8(r * 255), uint8(g * 255), uint8(b * 255), uint8(a * 255)}...)
	}
	img := image.RGBA{Pix: pixels, Stride: 1, Rect: image.Rectangle{image.Point{0, 0}, image.Point{len(pixels) / 4, 1}}}
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
	active       *int //The active tick
}

//GradientEditor creates a GradientEditor for use in immediate mode gui
func GradientEditor(id string, gradient *Gradient, currentColor *int, flags int) GradientEditorWidget {

	//img:=gradient.makePreview()
	g := GradientEditorWidget{}
	g.grad = gradient
	g.active = currentColor
	g.UpdateTex()
	return g
}

func arrToVec4(arr [4]float32) imgui.Vec4 {
	return imgui.Vec4{X: arr[0], Y: arr[1], Z: arr[2], W: arr[3]}
}

//Build provides the build function for immediate mode
func (gr GradientEditorWidget) Build() {
	avail := imgui.ContentRegionAvail()
	var availableWidth float32 = avail.X //100.0 //TODO make this take up the correct amount of space
	imgui.Text("Gradient")
	p := giu.GetCursorScreenPos()

	//imgui.ImageButton(gr.previewTexID, imgui.Vec2{availableWidth, 10})
	imgui.ImageButtonV(gr.previewTexID, imgui.Vec2{X: availableWidth, Y: 10}, imgui.Vec2{X: 0, Y: 0}, imgui.Vec2{X: 1, Y: 1}, 0, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1})
	if imgui.IsItemClicked(0) {
		fmt.Println("add color")
		//Add color at mouse position
		m := imgui.MousePos()
		xpercent := clamp(float64((m.X-float32(p.X))/availableWidth), 0.0, 1.0)
		c := gr.grad.GetColorAt(xpercent)
		gr.grad.AddTick(xpercent, c)
	}

	drawList := imgui.GetWindowDrawList()

	tickHeight := 14
	tickBgCol := giu.ToVec4Color(color.RGBA{uint8(28), uint8(36), uint8(43), uint8(0xff)})
	//imgui.Vec4{1,0,0,1}//
	tickBorder := 2
	topOffset := 10
	tickSize := imgui.Vec2{X: float32(tickBorder*2 + topOffset), Y: float32(tickHeight + topOffset + tickBorder*2)}
	for i, t := range gr.grad.ticks {
		col := arrToVec4(t.color) //giu.ToVec4Color(t.color)
		pos := float32(t.pos) * availableWidth

		if i == *gr.active {
			tickBgCol = giu.ToVec4Color(color.RGBA{uint8(71), uint8(143), uint8(255), uint8(0xff)})
		} else {
			tickBgCol = giu.ToVec4Color(color.RGBA{uint8(28), uint8(36), uint8(43), uint8(0xff)})

		}

		pmin := p.Add(image.Point{int(pos) - tickBorder, 10 - tickBorder})
		pmax := p.Add(image.Point{10 + int(pos) + tickBorder, tickHeight + 10 + tickBorder})
		drawList.AddRectFilled(giu.ToVec2(pmin), giu.ToVec2(pmax), tickBgCol, 0, 5)

		pmin2 := p.Add(image.Point{int(pos), 10})
		pmax2 := p.Add(image.Point{10 + int(pos), tickHeight + 10})
		drawList.AddRectFilled(giu.ToVec2(pmin2), giu.ToVec2(pmax2), col, 0, 5)

		imgui.SetCursorScreenPos(giu.ToVec2(p.Add(image.Point{int(availableWidth*float32(t.pos)) - tickBorder, 0})))
		if imgui.InvisibleButton("ButtonX"+fmt.Sprint(i), tickSize) {
			fmt.Println("clicked: ", i)
			*gr.active = i
			fmt.Println("Active: ", gr.active)
		}
		if imgui.IsItemActive() {
			io := imgui.CurrentIO()
			md := io.GetMouseDelta()
			gr.grad.SetPos(i, clamp(t.pos+float64(md.X/availableWidth), 0.0, 1.0))
			*gr.active = i
		}
		if imgui.IsItemClicked(2) {
			*gr.active = -1
		}
		if imgui.IsItemClicked(1) {
			*gr.active = -1
			fmt.Println("Deleting", i)
			gr.grad.RemoveTick(i)
		}
	}
	if *gr.active != -1 {
		imgui.ColorPicker4("# Gradient Color Picker", &gr.grad.ticks[*gr.active].color)
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
	if err != nil {
		fmt.Println("Texture Update Error ERR: ", err)
	}
}
