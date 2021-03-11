package main
import ("fmt"
	"image/color"
	"github.com/AllenDang/giu/imgui"

)
var pickers []ColorPickerWidget = []ColorPickerWidget{}

func BuildColorPickers(){
	for _,p := range(pickers){
		p.Build()
	}
}



func ColorPicker(id string, cols *[3]float32, flags int) ColorPickerWidget {
	p:=ColorPickerWidget{id, cols, flags}
	pickers=append(pickers, p)
	fmt.Println("pickers: ", pickers)
	return p
}
type ColorPickerWidget struct{
	id string
	cols *[3]float32
	flags int
}

func (c ColorPickerWidget) GetColor() color.RGBA {
	return colFromArr(*(c.cols))
}
func (c ColorPickerWidget) Build(){
	imgui.ColorPicker3(c.id, c.cols,c.flags)
}

func colFromArr(arr [3]float32 ) color.RGBA{
	return color.RGBA{uint8(arr[0]*255),uint8(arr[1]*255),uint8(arr[2]*255),0xff}
}

func makeColors() []color.RGBA {//,err
	cols:=[]color.RGBA{}
		for _,p := range(pickers){
		cols=append(cols, p.GetColor())
	}
	return cols //,err if there are no colors
}