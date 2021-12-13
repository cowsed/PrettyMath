package modelviewer

import (
	"fmt"

	g "github.com/AllenDang/giu"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Renderable interface {
	Draw(projection, view mgl32.Mat4, program uint32)
	Build() g.Widget
}

type LoadedModel struct {
	id            int
	filename      string
	Translation   Vec3Param
	Scale         float32
	DrawWireframe bool
	obj           objModel
	numpoints     int
	errortext     string
	vao           uint32
	vbo           uint32
	normalbuffer  uint32
}

func NewLoadedModel(id int) *LoadedModel {
	m := LoadedModel{
		id:       id,
		filename: "model.obj",
		Translation: Vec3Param{
			Name: "Translation ## " + fmt.Sprint(id),
			min:  -10000,
			max:  10000,
			step: .01},
		Scale: 1,
	}
	return &m
}
func (l *LoadedModel) Build() g.Widget {
	return g.TreeNode(fmt.Sprintf("Object ## %d", l.id)).Layout(

		g.InputText(fmt.Sprintf("Filename ## %d", l.id), &l.filename),
		&l.Translation,
		g.Checkbox("Draw Wireframe", &l.DrawWireframe),
		g.Button("Reload ##"+l.filename).OnClick(l.LoadFromFile),
	)
}

func (l *LoadedModel) Draw(projection, view mgl32.Mat4, program uint32) {
	modelMatrixName := "modelMatrix"
	viewMatrixName := "viewMatrix"
	projMatrixName := "projMatrix"
	mvpMatrixName := "MVP"

	// Set up model martix for shader
	model := mgl32.Ident4()

	model = mgl32.Translate3D(l.Translation.Value[0], l.Translation.Value[1], l.Translation.Value[2])

	model = model.Mul4(mgl32.Scale3D(l.Scale, l.Scale, l.Scale))

	// Set the modelUniform for the object
	modelUniform := gl.GetUniformLocation(program, gl.Str(modelMatrixName+"\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	// Set the viewUniform for the object
	viewUniform := gl.GetUniformLocation(program, gl.Str(viewMatrixName+"\x00"))
	gl.UniformMatrix4fv(viewUniform, 1, false, &view[0])

	// Set the projectionUniform for the object
	projectionUniform := gl.GetUniformLocation(program, gl.Str(projMatrixName+"\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	// Also pass the combined MVP uniform for convenience
	MVP := projection.Mul4(view.Mul4(model))
	MVPUniform := gl.GetUniformLocation(program, gl.Str(mvpMatrixName+"\x00"))
	gl.UniformMatrix4fv(MVPUniform, 1, false, &MVP[0])

	ShadeNormPos := gl.GetUniformLocation(program, gl.Str("ShadeNormal"+"\x00"))
	gl.Uniform1i(ShadeNormPos, 1)

	gl.BindVertexArray(l.vao)

	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(l.numpoints/3))

	//Draw Borders
	if l.DrawWireframe {
		gl.Uniform1i(ShadeNormPos, 0)
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(l.numpoints/3))
	}
}

func (l *LoadedModel) LoadFromFile() {
	l.errortext = ""

	model, err := readOBJ(l.filename)
	if err != nil {
		l.errortext = err.Error()
		return
	}
	l.obj = model
	points := model.ToArrayXYZ()
	normals := model.ToArrayNormals()
	l.numpoints = len(points)

	gl.GenBuffers(1, &l.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, l.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	gl.GenVertexArrays(1, &l.vao)
	gl.BindVertexArray(l.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, l.vbo)

	gl.GenBuffers(1, &l.normalbuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, l.normalbuffer)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(normals), gl.Ptr(normals), gl.STATIC_DRAW)

	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, l.vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	gl.EnableVertexAttribArray(1)
	gl.BindBuffer(gl.ARRAY_BUFFER, l.normalbuffer)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, true, 0, nil)

}
