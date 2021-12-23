package modelviewer

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	g "github.com/AllenDang/giu"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Renderable interface {
	Draw(projection, view mgl32.Mat4, program uint32)
	Build() g.Widget
	Dispose()
	SetID(id int)
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

func (l *LoadedModel) SetID(id int) {
	l.id = id
}
func (l *LoadedModel) Dispose() {
	gl.DeleteBuffers(1, &l.normalbuffer)
	gl.DeleteBuffers(1, &l.vbo)
	gl.DeleteVertexArrays(1, &l.vao)

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

type faceIndex struct {
	f1 []int32 // [v1, uv1, n1]
	f2 []int32 // [v2, uv2, n2]
	f3 []int32 // [v3, uv3, n3]
}

type objModel struct {
	meshName string
	vertices []mgl32.Vec3
	uvs      []mgl32.Vec2
	normals  []mgl32.Vec3
	faces    []faceIndex
}

func (m objModel) ToArrayXYZ() []float32 {
	var verticeArray []float32
	println("Num faces: ", len(m.faces))
	println("Num Pouints: ", len(m.vertices))
	for _, face := range m.faces {
		// Vertice 1
		v1 := m.vertices[face.f1[0]]
		verticeArray = append(verticeArray, v1.X(), v1.Y(), v1.Z())

		// Vertice 2
		v2 := m.vertices[face.f2[0]]
		verticeArray = append(verticeArray, v2.X(), v2.Y(), v2.Z())

		// Vertice 3
		v3 := m.vertices[face.f3[0]]
		verticeArray = append(verticeArray, v3.X(), v3.Y(), v3.Z())

	}
	fmt.Println("Verts sent: ", len(verticeArray))
	return verticeArray

}
func (m objModel) ToArrayNormals() []float32 {
	var normalArray []float32
	println("Num faces: ", len(m.faces))
	println("Num Normals: ", len(m.normals))
	for _, face := range m.faces {
		// Vertice 1
		n1 := m.normals[face.f1[2]]
		normalArray = append(normalArray, n1.X(), n1.Y(), n1.Z())

		// Vertice 2
		n2 := m.normals[face.f2[2]]
		normalArray = append(normalArray, n2.X(), n2.Y(), n2.Z())

		// Vertice 3
		n3 := m.normals[face.f3[2]]
		normalArray = append(normalArray, n3.X(), n3.Y(), n3.Z())

	}
	fmt.Println("Normals sent: ", len(normalArray))
	return normalArray

}

func readOBJ(filePath string) (objModel, error) {
	file, err := os.Open(filePath)

	if err != nil {
		return objModel{}, fmt.Errorf("failed opening obj file: %s", err)
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	var model objModel
	for fileScanner.Scan() {
		text := fileScanner.Text()
		values := strings.Split(text, " ")

		switch values[0] {
		case "o":
			// Mesh name
			model.meshName = values[1]
		case "v":
			// Vertice
			x, _ := strconv.ParseFloat(values[1], 32)
			y, _ := strconv.ParseFloat(values[2], 32)
			z, _ := strconv.ParseFloat(values[3], 32)
			model.vertices = append(model.vertices, mgl32.Vec3{float32(x), float32(y), float32(z)})
		case "vt":
			// uvs
			u, _ := strconv.ParseFloat(values[1], 32)
			v, _ := strconv.ParseFloat(values[2], 32)
			model.uvs = append(model.uvs, mgl32.Vec2{float32(u), float32(v)})
		case "vn":
			// Vertice normal
			x, _ := strconv.ParseFloat(values[1], 32)
			y, _ := strconv.ParseFloat(values[2], 32)
			z, _ := strconv.ParseFloat(values[3], 32)
			model.normals = append(model.normals, mgl32.Vec3{float32(x), float32(y), float32(z)})
		case "f":
			// face indices
			// e.g. 24/33/37 31/28/37 37/47/37
			f1text := strings.Split(values[1], "/")
			f2text := strings.Split(values[2], "/")
			f3text := strings.Split(values[3], "/")

			var face faceIndex
			// -1 on final index since obj indexing starts at 1 (we want 0)
			fv1, _ := strconv.ParseInt(f1text[0], 10, 32)
			fuv1, _ := strconv.ParseInt(f1text[1], 10, 32)
			fn1, _ := strconv.ParseInt(f1text[2], 10, 32)
			face.f1 = append(face.f1, int32(fv1)-1, int32(fuv1)-1, int32(fn1)-1)

			fv2, _ := strconv.ParseInt(f2text[0], 10, 32)
			fuv2, _ := strconv.ParseInt(f2text[1], 10, 32)
			fn2, _ := strconv.ParseInt(f2text[2], 10, 32)
			face.f2 = append(face.f2, int32(fv2)-1, int32(fuv2)-1, int32(fn2)-1)

			fv3, _ := strconv.ParseInt(f3text[0], 10, 32)
			fuv3, _ := strconv.ParseInt(f3text[1], 10, 32)
			fn3, _ := strconv.ParseInt(f3text[2], 10, 32)
			face.f3 = append(face.f3, int32(fv3)-1, int32(fuv3)-1, int32(fn3)-1)

			model.faces = append(model.faces, face)
		}

	}

	return model, nil
}
