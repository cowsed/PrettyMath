package modelviewer

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

//Draws the shader
func (ws *Workspace) Draw() {

	if ws.doturntable {
		ws.theta += 0.01
		ws.cameraPos.Value = mgl32.Rotate3DY(ws.theta).Mul3x1(mgl32.Vec3{ws.radius, 0, 0}).Add(ws.lookatPos.Value)

		ws.cameraPos.Value[1] = ws.height
	}
	// Set up projection matrix for shader
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(1920)/1080, 0.1, 40.0)

	// Set up view matrix for shader
	view := mgl32.LookAtV(ws.cameraPos.Value, ws.lookatPos.Value, mgl32.Vec3{0, 1, 0})

	//Actually draw the tris

	start := time.Now()

	//Switch to shader fbo
	gl.BindFramebuffer(gl.FRAMEBUFFER, ws.framebuffer)
	gl.Viewport(0, 0, 1920, 1080)

	//PRoblem seems to be in the depth func not getting set correctyl
	//possible solution is to make the workspace have a opengl draw function so main can call it after imgui rendering is over so they dont overlap
	//just a guess tho and that may involve wackiness because of giu
	gl.ClearColor(ws.clearcolor[0], ws.clearcolor[1], ws.clearcolor[2], ws.clearcolor[3])
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
	gl.DepthFunc(gl.LESS)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.UseProgram(ws.program)

	lightpospos := gl.GetUniformLocation(ws.program, gl.Str("lightpos"+"\x00"))

	gl.Uniform3f(lightpospos, ws.lightpos.Value[0], ws.lightpos.Value[1], ws.lightpos.Value[2])

	ambientPos := gl.GetUniformLocation(ws.program, gl.Str("ambient"+"\x00"))
	gl.Uniform1f(ambientPos, ws.ambientLight)

	ShadeNormPos := gl.GetUniformLocation(ws.program, gl.Str("ShadeNormal"+"\x00"))
	gl.Uniform1i(ShadeNormPos, 1)

	for i := 0; i < len(ws.Objects); i++ {
		ws.Objects[i].Draw(projection, view, ws.program)
	}

	//Switch back to default framebuffer
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	elapsed := time.Since(start)
	ws.lastTime = float64(elapsed.Microseconds())

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
	fragmentShader, err := ws.compileShader(baseFragSource+"\x00", gl.FRAGMENT_SHADER)
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

		return 0, fmt.Errorf("failed to compile:\nLog:\n%v", log[:len(log)-2])
	}
	return shader, nil
}
