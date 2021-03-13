package tools

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
	//Sets two values black at 0, white at 1
	g.AddTick(0.0, color.RGBA{0, 0, 0, 0xff})
	g.AddTick(0.3, color.RGBA{255, 0, 255, 0xff})
	g.AddTick(0.6, color.RGBA{255, 255, 255, 0xff})
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

func lerp(a, b uint8, amt float64) uint8 {
	return uint8(float64(a)*(amt) + float64(b)*(1-amt))
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
func p2V2(p image.Point) imgui.Vec2{
	return imgui.Vec2{float32(p.X),float32(p.Y)}
}
func (gr GradientEditorWidget) Build() {
	//w,h:=
	var availableWidth float32 =100.0
	imgui.Text("Gradient")
	p:=giu.GetCursorScreenPos()

	imgui.ImageButton(gr.previewTexID, imgui.Vec2{availableWidth, 10})
	drawList:=imgui.GetWindowDrawList()
	
	tickHeight:=14
	tickBgCol:=giu.ToVec4Color(color.RGBA{uint8(28),uint8(36),uint8(43),uint8(0xff)})
	//imgui.Vec4{1,0,0,1}//
	tickBorder:=2
	topOffset:=10
	tickSize:=imgui.Vec2{float32(tickBorder*2+topOffset), float32(tickHeight+topOffset+tickBorder*2)}
	for i,t := range(gr.grad.ticks){
		col:=giu.ToVec4Color(t.color)
		pos:=float32(t.pos)*availableWidth
		
		pmin:=p.Add(image.Point{int(pos)-tickBorder,10-tickBorder})
		pmax:=p.Add(image.Point{10+int(pos)+tickBorder,tickHeight+10+tickBorder})
		drawList.AddRectFilled(p2V2(pmin), p2V2(pmax), tickBgCol,0,5)
	
		pmin2:=p.Add(image.Point{int(pos),10})
		pmax2:=p.Add(image.Point{10+int(pos),tickHeight+10})
		drawList.AddRectFilled(p2V2(pmin2), p2V2(pmax2), col,0,5)
		
		imgui.SetCursorScreenPos(p2V2( p.Add(image.Point{int(availableWidth*float32(t.pos)),0} ) ))
		if imgui.ButtonV("ButtonX"+string(i), tickSize){
			fmt.Println("clicked: ",i)
		}
		if imgui.IsItemHovered(){
			fmt.Println("Hovered")
		}
	}
}
func (gr *GradientEditorWidget) UpdateTex() {
	renderer := giu.Context.GetRenderer()
	img := gr.grad.makePreview()
	//Discard old texture
	renderer.ReleaseImage(gr.previewTexID)

	texid, err := renderer.LoadImage(img)
	gr.previewTexID = texid
	fmt.Println("Texture Update Error ERR: ", err)

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

