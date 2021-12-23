package modelviewer

import (
	_ "embed"
	"fmt"

	"github.com/AllenDang/giu/imgui"
	"github.com/chewxy/math32"
	tools "github.com/cowsed/PrettyMath/Tools"
	workspace "github.com/cowsed/PrettyMath/Workspaces"

	"github.com/AllenDang/giu"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

//Register To Workspaces
func init() {
	workspace.RegisterWorkspace(Init, "Model Viewer")
}

//go:embed Shaders/example.frag
var baseFragSource string

//go:embed Shaders/example.vert
var baseVertSource string

type Workspace struct {
	amOpen  bool
	onClose func()

	cameraPos    Vec3Param
	lookatPos    Vec3Param
	lightpos     Vec3Param
	ambientLight float32
	imageZoom    float32
	lastTime     float64

	theta       float32 //for doing turntable camera
	radius      float32
	height      float32
	doturntable bool

	clearcolor []float32

	//Opengl stuff
	program     uint32
	framebuffer uint32
	outputTex   uint32

	NewObjectType int32
	Objects       []Renderable
}

//Init initializes a new plotter workspace
func Init(onCloseFunc func(), AddProcessComm func() chan workspace.ProgressUpdate) tools.Workspace {
	ws := Workspace{
		amOpen:  true,
		onClose: onCloseFunc,
		cameraPos: Vec3Param{
			Value: [3]float32{8, 6, 0},
			Name:  "Camera Position",
			min:   -10000,
			max:   10000,
			step:  .01,
		},
		lookatPos: Vec3Param{
			Value: [3]float32{},
			Name:  "Center Position",
			min:   -10000,
			max:   10000,
			step:  .01,
		},
		lightpos: Vec3Param{
			Value: [3]float32{1, -1, 0},
			Name:  "Light Position",
			min:   -10000,
			max:   10000,
			step:  .01,
		},
		imageZoom:    1,
		ambientLight: .7,
		clearcolor:   []float32{1, 1, 1, 1},
		Objects:      []Renderable{},
	}
	//mod := NewLoadedModel(0)
	//mod.filename = "Workspaces/ModelViewer/Models/monkey.obj"
	//mod.LoadFromFile()
	mod := NewSquareGenerator(0)
	mod.spinePoints = 20

	mod.Function = "2+.5*cos(3.14159*x)"
	mod.DrawWireframe = true
	mod.CreateMesh()

	ws.Objects = append(ws.Objects, mod)

	//OpenGL Stuff =========

	//Generate new framebuffer
	gl.GenFramebuffers(1, &ws.framebuffer)

	fmt.Println("Made fbo", ws.framebuffer)
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		fmt.Println("FB making error")
	}

	//Rendered Texture
	gl.GenTextures(1, &ws.outputTex)
	gl.BindTexture(gl.TEXTURE_2D, ws.outputTex)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1920, 1080, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)

	gl.BindFramebuffer(gl.FRAMEBUFFER, ws.framebuffer)
	gl.FramebufferTexture(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, ws.outputTex, 0)

	var rbo uint32
	gl.CreateRenderbuffers(1, &rbo)

	gl.BindRenderbuffer(gl.RENDERBUFFER, rbo)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH24_STENCIL8, 1920, 1080)
	gl.BindRenderbuffer(gl.RENDERBUFFER, 0)

	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.RENDERBUFFER, rbo)

	//Bind back the default framebuffer
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	ws.BuildProgram()

	return &ws
}

func (ws *Workspace) CreateNewObject() {
	var newObject Renderable
	switch ws.NewObjectType {
	case 0:
		newObject = NewLoadedModel(len(ws.Objects))
	case 1:
		newObject = NewDiskGenerator(len(ws.Objects))
	case 2:
		newObject = NewSquareGenerator(len(ws.Objects))
	}
	ws.Objects = append(ws.Objects, newObject)
}

func (ws *Workspace) ToggleTurntable() {
	if !ws.doturntable {
		return
	}
	fmt.Println(ws.cameraPos.Value)
	ws.height = ws.cameraPos.Value[1]
	v := mgl32.Vec2{ws.cameraPos.Value[0] - ws.lookatPos.Value[0], ws.cameraPos.Value[2] - ws.lookatPos.Value[2]}
	ws.radius = v.Len()
	ws.theta = -math32.Atan2(v[1], v[0])
	fmt.Println(ws.radius, ws.theta, ws.height)
}

func (ws *Workspace) Build() {
	ws.Draw()

	//Close when necessary
	if !ws.amOpen {
		println("Closing\n\n\n\n\n\n\n")
		for i := range ws.Objects {
			ws.Objects[i].Dispose()
		}
		//Delete gl stuff
		//Still a bit of a problem here when freeing as it seems not everything is freed
		gl.DeleteFramebuffers(1, &ws.framebuffer)
		gl.DeleteTextures(1, &ws.outputTex)
		gl.DeleteProgram(ws.program)
		ws.onClose()
	}

	cameraControls := giu.Group().Layout(
		&ws.cameraPos,
		giu.Checkbox("Turntable: ", &ws.doturntable).OnChange(ws.ToggleTurntable),
		&ws.lookatPos,
	)

	Adder := giu.Combo("Create new Object", "From File", []string{"From File", "Function Defined Revolution", "Function Defined Square"}, &ws.NewObjectType).OnChange(ws.CreateNewObject)
	var ObjectsUI giu.Widget = giu.Group()

	if len(ws.Objects) > 0 {
		widgets := make([]giu.Widget, len(ws.Objects))
		for i := 0; i < len(ws.Objects); i++ {
			widgets[i] = ws.Objects[i].Build()
		}
		ObjectsUI = giu.TreeNode("Objects").Layout(widgets...)

	}

	giu.TabItem("ModelViewer").Layout(
		giu.SplitLayout("MainSplit", giu.DirectionHorizontal, true, 300,
			giu.Group().Layout(
				cameraControls,
				giu.Label(fmt.Sprintf("ModelViewer Framebuf: %d. OutputTex: %d", ws.framebuffer, ws.outputTex)),
				giu.Custom(func() {
					tools.DragFloatN("Clear Color", ws.clearcolor, 0.01, 0, 1, "%f")

				}),
				&ws.lightpos,
				giu.SliderFloat("Ambient Light", &ws.ambientLight, 0, 1),
				ObjectsUI,
				Adder),
			giu.Group().Layout(
				giu.Label(fmt.Sprintf("Draw Time: %.2f us ", ws.lastTime)),
				giu.Custom(func() {
					size := imgui.ContentRegionAvail()
					w := size.X
					var aspect float32 = 1920.0 / 1080.0
					size2 := imgui.Vec2{X: w * ws.imageZoom, Y: w / aspect * ws.imageZoom}
					imgui.ImageV(imgui.TextureID(ws.outputTex), size2, imgui.Vec2{X: 0, Y: 1}, imgui.Vec2{X: 1, Y: 0}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}, imgui.Vec4{})
					//imgui.Image(imgui.TextureID(ws.outputTex))
				}),
			),
		),
	).IsOpen(&ws.amOpen).Build()

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
