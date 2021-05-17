package OpenGLEditor

import (
	"fmt"

	"image"
	//Enable formats

	"image/draw"
	_ "image/jpeg"
	_ "image/png"

	"os"

	"github.com/AllenDang/giu/imgui"
	"github.com/go-gl/gl/v3.2-core/gl"
)

func (ws *Workspace) FindUniforms() {
	var i uint32
	var count int32

	var size int32   // size of the variable
	var uType uint32 // type of the variable (float, vec3 or mat4, etc)

	const bufSize int32 = 32                 // maximum name length
	var uName []byte = make([]byte, bufSize) // variable name in GLSL
	var length int32                         // name length

	gl.GetProgramiv(ws.program, gl.ACTIVE_UNIFORMS, &count)
	newUniforms := make([]Parameter, count)

	for i = 0; i < uint32(count); i++ {
		gl.GetActiveUniform(ws.program, i, bufSize, &length, &size, &uType, &uName[0])

		fmt.Printf("Name: %v Type: %v\n", gl.GoStr(&uName[0]), uType)
		//Reuse old one if still applicable
		if len(ws.parameters) == len(newUniforms) && ws.parameters[i].GetName() == gl.GoStr(&uName[0]) {
			newUniforms[i] = ws.parameters[i]
			continue
		}
		switch uType {
		case gl.SAMPLER_2D:
			newUniforms[i] = &Sampler2DParam{
				path:      "image/path",
				img:       nil,
				texHandle: 0,
				name:      gl.GoStr(&uName[0]),
				status:    "Image Unloaded",
			}
		case gl.FLOAT:

			newUniforms[i] = &FloatParam{
				Value: 0,
				Name:  gl.GoStr(&uName[0]),
				min:   -1000,
				max:   1000,
				step:  .05,
			}
		case gl.FLOAT_VEC3:
			newUniforms[i] = &Vec3Param{
				Value: [3]float32{0, 0, 0},
				Name:  gl.GoStr(&uName[0]),
				min:   -1000,
				max:   1000,
				step:  .05}
		case gl.INT:
			newUniforms[i] = &Int32Param{
				Name:  gl.GoStr(&uName[0]),
				Value: 0,
			}
		}

	}
	ws.parameters = newUniforms
}

type Parameter interface {
	Build()
	SetUniform(program uint32)
	GetName() string
}

type Sampler2DParam struct {
	path      string
	img       *image.RGBA
	texHandle uint32
	name      string
	status    string
}

func (p *Sampler2DParam) SetUniform(program uint32) {
	loc := gl.GetUniformLocation(program, gl.Str(p.name+"\x00"))
	gl.Uniform1ui(loc, p.texHandle)
}

func (p *Sampler2DParam) GetName() string {
	return p.name
}
func (p *Sampler2DParam) Build() {
	imgui.InputText(p.name+"filename", &p.path)
	imgui.SameLine()

	if imgui.BeginPopupContextItemV(p.name+" Sampler2d", 1) {
		imgui.BeginGroup()
		id := "var: " + p.name + " : Sampler 2D"
		imgui.Text(id)
		imgui.Text("Further information and control")
		imgui.Text(fmt.Sprintf("OpenGL ID %d", p.texHandle))
		if p.img != nil {
			imgui.Text(fmt.Sprintf("%dx%d", p.img.Bounds().Dx(), p.img.Bounds().Dy()))
			imgui.Image(imgui.TextureID(p.texHandle), imgui.Vec2{X: 400, Y: 400 * (float32(p.img.Bounds().Dy()) / float32(p.img.Bounds().Dx()))})
		} else {
			imgui.Text("No GL Texture yet")
		}
		imgui.EndGroup()
		imgui.EndPopup()
	}

	if imgui.Button("Reload") {
		p.Reload()
	}

	if p.status != "" {

		imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: 1, Y: 0, Z: 0, W: 1})
		imgui.Text(p.status)
		imgui.PopStyleColor()
	}
}

func (p *Sampler2DParam) Reload() {
	//Delete old handle
	if p.texHandle != 0 {
		gl.DeleteTextures(1, &p.texHandle)
		p.texHandle = 0
	}

	//Reloads the file
	f, err := os.Open(p.path)
	if err != nil {
		p.status = err.Error()
		return
	}
	defer f.Close()

	//Create Temporary image.Image
	var fmtstr string
	var tmpImg image.Image
	tmpImg, fmtstr, err = image.Decode(f)
	if err != nil {
		p.status = err.Error()
		return
	}
	fmt.Println("Loaded Format:", fmtstr)

	//Create actual image and draw to it
	p.img = image.NewRGBA(tmpImg.Bounds())
	draw.Draw(p.img, p.img.Bounds(), tmpImg, image.Pt(0, 0), 0)

	//Do the opengl stuff
	gl.GenTextures(1, &p.texHandle)
	gl.BindTexture(gl.TEXTURE_2D, p.texHandle)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(p.img.Bounds().Dx()), int32(p.img.Bounds().Dy()), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(p.img.Pix))
	gl.GenerateMipmap(gl.TEXTURE_2D)

	p.status = ""

}

type Int32Param struct {
	Value int32
	Name  string
}

func (b *Int32Param) GetName() string {
	return b.Name
}
func (b *Int32Param) Build() {
	imgui.DragInt(b.Name, &b.Value)
}
func (b *Int32Param) SetUniform(program uint32) {
	loc := gl.GetUniformLocation(program, gl.Str(b.Name+"\x00"))
	gl.Uniform1i(loc, b.Value)
}

type FloatParam struct {
	Value          float32
	Name           string
	min, max, step float32
}

func (fp *FloatParam) GetName() string {
	return fp.Name
}
func (fp *FloatParam) Build() {
	imgui.DragFloatV(fp.Name, &fp.Value, fp.step, fp.min, fp.max, "%.3f", 1)

	if imgui.BeginPopupContextItemV(fp.Name+" float", 1) {
		imgui.BeginGroup()
		id := "var: " + fp.Name + " : Float"
		imgui.Text(id)
		imgui.Text("Further information and control")
		imgui.InputFloat("min", &fp.min)
		imgui.InputFloat("max", &fp.max)
		imgui.InputFloat("step", &fp.step)

		imgui.EndGroup()

		imgui.EndPopup()
	}
}
func (fp *FloatParam) SetUniform(program uint32) {
	loc := gl.GetUniformLocation(program, gl.Str(fp.Name+"\x00"))
	gl.Uniform1f(loc, fp.Value)
}

type Vec3Param struct {
	Value          [3]float32
	Name           string
	min, max, step float32
}

func (v3p *Vec3Param) GetName() string {
	return v3p.Name
}
func (v3p *Vec3Param) Build() {
	DragFloat3(v3p.Name, &v3p.Value, v3p.step, v3p.min, v3p.max, "%.3f")

	if imgui.BeginPopupContextItemV(v3p.Name+" Vec3", 1) {
		imgui.BeginGroup()
		id := "var: " + v3p.Name + " : Vec3"
		imgui.Text(id)
		imgui.Text("Further information and control")
		imgui.InputFloat("min", &v3p.min)
		imgui.InputFloat("max", &v3p.max)
		imgui.InputFloat("step", &v3p.step)

		imgui.EndGroup()

		imgui.EndPopup()
	}
}
func (fp *Vec3Param) SetUniform(program uint32) {
	loc := gl.GetUniformLocation(program, gl.Str(fp.Name+"\x00"))
	gl.Uniform3f(loc, fp.Value[0], fp.Value[1], fp.Value[2])
}

func DragFloat3(label string, vec *[3]float32, speed, min, max float32, format string) bool {
	value_changed := false
	size := imgui.CalcItemWidth() / float32(len(vec)+1)
	for i := range vec {
		imgui.PushItemWidth(size)
		id := fmt.Sprintf("%s-%d\n", label, i)
		imgui.PushID(id)
		if i > 0 {
			imgui.SameLine()
		}
		changed := imgui.DragFloatV("", &vec[i], speed, min, max, format, 0)
		value_changed = value_changed || changed
		imgui.PopID()
		imgui.PopItemWidth()
	}

	imgui.SameLine()
	imgui.Text(label)

	//imgui.EndGroup()
	return value_changed
}
