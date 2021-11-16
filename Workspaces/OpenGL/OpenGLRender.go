package OpenGLEditor

import (
	_ "embed"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"strings"
	"time"
	"unsafe"

	"github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/go-gl/gl/v3.2-core/gl"
)

var points []float32 = []float32{
	-1, 1, 0,
	-1, -1, 0,
	1, -1, 0,

	-1, 1, 0,
	1, 1, 0,
	1, -1, 0,
}

//go:embed example.frag
var baseFragSource string

//go:embed example.vert
var baseVertSource string

func Init(onCloseFunc func()) Workspace {

	ws := Workspace{
		amOpen:     true,
		onClose:    onCloseFunc,
		imageZoom:  .5,
		fragSource: baseFragSource,
		parameters: []Parameter{},
	}
	{
		//Sets up editor
		ws.errMarkers = imgui.NewErrorMarkers()
		ws.editor = imgui.NewTextEditor()
		ws.editor.SetErrorMarkers(ws.errMarkers)
		ws.editor.SetLanguageDefinitionC()
		ws.editor.SetText(ws.fragSource)
		ws.editor.SetShowWhitespaces(false)
		ws.editor.SetTabSize(4)
	}
	gl.GenBuffers(1, &ws.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, ws.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	gl.GenVertexArrays(1, &ws.vao)
	gl.BindVertexArray(ws.vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, ws.vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

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

	// Set the list of draw buffers.
	DrawBuffers := []uint32{gl.COLOR_ATTACHMENT0}
	gl.DrawBuffers(1, &DrawBuffers[0]) // "1" is the size of DrawBuffers
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	ws.BuildProgram()

	return ws
}

type Workspace struct {
	amOpen     bool
	onClose    func()
	fragSource string
	imageZoom  float32

	lastTime float64

	//Shader Parameters
	parameters []Parameter

	//Opengl stuff
	program     uint32
	framebuffer uint32
	outputTex   uint32
	vao         uint32
	vbo         uint32

	editor     imgui.TextEditor
	errMarkers imgui.ErrorMarkers
}

//Draws the shader
func (ws *Workspace) Draw() {
	start := time.Now()
	ws.SetUniforms()
	//Switch to shader fbo
	gl.BindFramebuffer(gl.FRAMEBUFFER, ws.framebuffer)
	gl.Viewport(0, 0, 1920, 1080)

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.ClearColor(1, 0, 1, 1)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	gl.UseProgram(ws.program)
	gl.BindVertexArray(ws.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(points)/3))

	//Switch back to default framebuffer
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	elapsed := time.Since(start)
	ws.lastTime = float64(elapsed.Microseconds())

}

//Sets the uniforms of the current shader program
func (ws *Workspace) SetUniforms() {
	for i := range ws.parameters {
		if ws.parameters[i] != nil {
			ws.parameters[i].SetUniform(ws.program)
		} else {
			fmt.Println("Unsuported Uniform type for uniform", i)
		}
	}
}

//Builds the opengl program
func (ws *Workspace) BuildProgram() {
	fmt.Println("Building")
	//Delete old program
	gl.DeleteProgram(ws.program)

	//Compile Vertex Shader
	vertexShader, err := ws.compileShader(baseVertSource+"\x00", gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}
	//Compile Fragment Shader
	fragmentShader, err := ws.compileShader(ws.fragSource+"\x00", gl.FRAGMENT_SHADER)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//Create and Link Program
	ws.program = gl.CreateProgram()
	gl.AttachShader(ws.program, vertexShader)
	gl.AttachShader(ws.program, fragmentShader)
	gl.LinkProgram(ws.program)

	//Release programs
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	//Check Link Errors
	var isLinked int32
	gl.GetProgramiv(ws.program, gl.LINK_STATUS, &isLinked)
	if isLinked == gl.FALSE {
		var maxLength int32
		gl.GetProgramiv(fragmentShader, gl.INFO_LOG_LENGTH, &maxLength)

		infoLog := make([]uint8, maxLength+1) //[bufSize]uint8{}
		gl.GetShaderInfoLog(fragmentShader, maxLength, &maxLength, &infoLog[0])

		fmt.Println("Link Infolog{", string(infoLog), "}")
		return
	}
	fmt.Println("Program", ws.program)
	//Find the uniforms
	ws.FindUniforms()
}
func (ws *Workspace) Build() {
	if !ws.amOpen {
		//Release Everything
		ws.onClose()
	}
	//Redraw Shader (Should realistically be controlled by a ticker or something)
	ws.Draw()

	giu.TabItem("OpenGL Editor").Layout(
		giu.Line(
			giu.Button("Build").OnClick(ws.BuildProgram),
			giu.Button("Save").OnClick(ws.SaveBuf),
			giu.SliderFloat("Image Size", &ws.imageZoom, 0, 1),
		),
		giu.SplitLayout("MainSplit", giu.DirectionHorizontal, true, 500,
			giu.Group().Layout(
				giu.Custom(func() {
					if imgui.TreeNodeV("Paramaters", imgui.TreeNodeFlagsFramed) {
						for i := range ws.parameters {
							ws.parameters[i].Build()
						}
						if len(ws.parameters) == 0 {
							giu.Label("You may have parameters that are not active")
						}
						imgui.TreePop()
					}
					ws.editor.Render("OpenCl", imgui.Vec2{X: 0, Y: 0}, true)
					if ws.editor.IsTextChanged() {
						ws.fragSource = ws.editor.GetText()
					}
				}),
			),
			giu.Group().Layout(
				giu.Label(fmt.Sprintf("Draw Time: %.2f us ", ws.lastTime)),
				//giu.Image(giu.ToTexture(imgui.TextureID(ws.outputTex))),
				giu.Custom(func() {
					imgui.Image(imgui.TextureID(ws.outputTex), imgui.Vec2{X: 1024 * ws.imageZoom, Y: 768 * ws.imageZoom})
				}),
			),
		),
	).IsOpen(&ws.amOpen).Build()
}

//SaveBuf somewhat of a debug function, saves the ID Framebuffer texture to a file
func (ws *Workspace) SaveBuf() {
	log.Println("Saving to a.png")
	outImage := image.NewRGBA(image.Rect(0, 0, 1920, 1080))
	fmt.Println(outImage.Pix[0:8])

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, ws.outputTex)

	gl.GetTexImage(gl.TEXTURE_2D,
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		unsafe.Pointer(&outImage.Pix[0]))
	fmt.Println(outImage.Pix[0:8])
	f, _ := os.Create("a.png")
	png.Encode(f, outImage)
}

//Compiles shaders
func (ws *Workspace) compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		ws.CreateAndSetErrors(log[:len(log)-2])
		return 0, fmt.Errorf("failed to compile:\nLog:\n%v", log[:len(log)-2])
	}
	ws.editor.SetErrorMarkers(imgui.NewErrorMarkers())
	return shader, nil
}

func (ws *Workspace) CreateAndSetErrors(log string) {
	eMarkers := imgui.NewErrorMarkers()
	errors := strings.Split(strings.Trim(log, "\x00\n"), "\n")
	for _, err := range errors {
		//skip EOF
		if err == "\x00" {
			continue
		}
		var text string
		var line, col int
		//This is really hacky cuz im bad at scanf

		err = strings.ReplaceAll(err, " ", "\\")
		fmt.Println("LineSource", err)
		n, e := fmt.Sscanf(err+"\n", "0:%d(%d):\\%s\n", &line, &col, &text)
		fmt.Println(n, "found")
		if e != nil {
			fmt.Println("ERR", n, "err:", e)
		}
		text = strings.ReplaceAll(text, "\\", " ")
		fmt.Printf("LineNum: %d, Col: %d, Result: %s\n", line, col, text)
		eMarkers.Insert(line, text)
	}

	ws.editor.SetErrorMarkers(eMarkers)

}
