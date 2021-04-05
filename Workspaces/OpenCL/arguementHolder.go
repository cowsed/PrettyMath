package opencl_renderer

import (
	"unsafe"

	"github.com/AllenDang/giu/imgui"
	"github.com/jgillich/go-opencl/cl"
)

//Type interface
type ClDataHolder interface {
	Build(ws *Workspace)
	SetArg(index int, k *cl.Kernel, ws *Workspace) error
	getType() string
	getName() string
	setName(string)
}

//Float input
type CLFloatInput struct {
	value          float32
	min, max, step float32
	name           string
	typeName       string
}

func (dh *CLFloatInput) Build(ws *Workspace) {
	//Maybe add tooltip that can be used to control secondary parameters
	//Last arguement is wonky because =the imgui package is slightly oout of dat and 1<<5 is loarytmic scale
	if imgui.DragFloatV(dh.name+" : "+dh.typeName, &dh.value, dh.step, dh.min, dh.max, "%.4f", 0) {
		ws.Run()
	}
	if imgui.BeginPopupContextItemV(dh.name+" "+dh.typeName, 1) {
		imgui.BeginGroup()
		id := "var: " + dh.name + " : " + dh.typeName
		imgui.Text(id)
		imgui.Text("Further information and control")
		imgui.InputFloat("min", &dh.min)
		imgui.InputFloat("max", &dh.max)
		imgui.InputFloat("step", &dh.step)

		imgui.EndGroup()

		imgui.EndPopup()
	}
}
func (dh *CLFloatInput) SetArg(index int, k *cl.Kernel, ws *Workspace) error {
	err := k.SetArgFloat32(index, dh.value)
	return err
}
func (dh *CLFloatInput) getType() string {
	return dh.typeName
}
func (dh *CLFloatInput) setName(n string) {
	dh.name = n
}
func (dh *CLFloatInput) getName() string {
	return dh.name
}

//Vector of Floats input
type CLFloatVecInput struct {
	values         []float32
	size           int
	min, max, step float32
	name           string
	typeName       string
}

func (dh *CLFloatVecInput) Build(ws *Workspace) {
	//Maybe add tooltip that can be used to control secondary parameters
	//Last arguement is wonky because =the imgui package is slightly oout of dat and 1<<5 is loarytmic scale
	//DragFloatN("TestN", data, 0.01,-10,10,"%.3f")
	if DragFloatN(dh.name+" : "+dh.typeName, dh.values, dh.step, dh.min, dh.max, "%.4f") {
		ws.Run()
	}
	if imgui.BeginPopupContextItemV(dh.name+" "+dh.typeName, 1) {
		imgui.BeginGroup()
		id := "var: " + dh.name + " : " + dh.typeName
		imgui.Text(id)
		imgui.Text("Further information and control")
		imgui.InputFloat("min", &dh.min)
		imgui.InputFloat("max", &dh.max)
		imgui.InputFloat("step", &dh.step)

		imgui.EndGroup()

		imgui.EndPopup()
	}
}
func (dh *CLFloatVecInput) SetArg(index int, k *cl.Kernel, ws *Workspace) error {
	//Size is 4*size float32 is 4 bytes * size
	//arr:=[3]float32{}

	err := k.SetArgUnsafe(index, 4*dh.size, unsafe.Pointer(&dh.values[0]))
	return err
}
func (dh *CLFloatVecInput) getType() string {
	return dh.typeName
}
func (dh *CLFloatVecInput) setName(n string) {
	dh.name = n
}
func (dh *CLFloatVecInput) getName() string {
	return dh.name
}

//int32 input
//Holds int information
type CLInt32Input struct {
	value    int32
	name     string
	min, max int32
	typeName string
}

func (dh *CLInt32Input) Build(ws *Workspace) {
	//Maybe add tooltip that can be used to control secondary parameters
	if imgui.DragIntV(dh.name+" : "+dh.typeName, &dh.value, 1, dh.min, dh.max, "%d") {
		ws.Run()
	}

	if imgui.BeginPopupContextItemV(dh.name+" "+dh.typeName, 1) {
		imgui.InputInt("min", &dh.min)
		imgui.InputInt("max", &dh.max)
		imgui.EndPopup()
	}
}
func (dh *CLInt32Input) SetArg(index int, k *cl.Kernel, ws *Workspace) error {
	err := k.SetArgInt32(index, int32(dh.value))
	return err
}
func (dh *CLInt32Input) getType() string {
	return dh.typeName
}
func (dh *CLInt32Input) getName() string {
	return dh.name
}
func (dh *CLInt32Input) setName(n string) {
	dh.name = n
}

//Uint32 input
//Holds uint information
type CLUint32Input struct {
	value    int32
	name     string
	min, max int32
	typeName string
}

func (dh *CLUint32Input) Build(ws *Workspace) {
	//Maybe add tooltip that can be used to control secondary parameters
	if imgui.DragIntV(dh.name+" : "+dh.typeName, &dh.value, 1, dh.min, dh.max, "%d") {
		ws.Run()
	}

	if imgui.BeginPopupContextItemV(dh.name+" "+dh.typeName, 1) {
		imgui.InputInt("min", &dh.min)
		imgui.InputInt("max", &dh.max)
		imgui.EndPopup()
	}
}
func (dh *CLUint32Input) SetArg(index int, k *cl.Kernel, ws *Workspace) error {
	err := k.SetArgUint32(index, uint32(dh.value))
	return err
}
func (dh *CLUint32Input) getType() string {
	return dh.typeName
}
func (dh *CLUint32Input) getName() string {
	return dh.name
}
func (dh *CLUint32Input) setName(n string) {
	dh.name = n
}

//Image buffer input
//Holds a reference to an image buffer with options to coordinate between kernels
type CLImageInput struct {
	identifier string
	index      int
	name       string
	show       bool
	typeName   string
}

func (dh *CLImageInput) Build(ws *Workspace) {
	imgui.BeginGroup()
	//Maybe add tooltip that can be used to control secondary parameters
	imgui.PushItemWidth(-1)
	imgui.Checkbox("##Show", &dh.show)
	imgui.PopItemWidth()
	if imgui.IsItemHovered() {
		imgui.SetTooltip("Display this image at the end")
	}
	imgui.SameLine()
	if imgui.InputText(dh.name, &dh.identifier) {
		ws.Run()
	}
	imgui.EndGroup()
	if imgui.BeginPopupContextItemV(dh.name+" "+dh.typeName, 1) {

		imgui.Text("More information")
		imgui.EndPopup()
	}
}
func (dh *CLImageInput) SetArg(index int, k *cl.Kernel, ws *Workspace) error {
	err := k.SetArgBuffer(index, ws.imageBuffers[dh.index])
	return err
}
func (dh *CLImageInput) getType() string {
	return dh.typeName
}
func (dh *CLImageInput) getName() string {
	return dh.name
}
func (dh *CLImageInput) setName(n string) {
	dh.name = n
}
