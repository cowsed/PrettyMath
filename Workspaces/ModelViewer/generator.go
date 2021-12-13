package modelviewer

import (
	"fmt"
	"image/color"
	"math"

	g "github.com/AllenDang/giu"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"

	parser "github.com/cowsed/Parser"
)

type DiskRevolution struct {
	id            int
	Translation   Vec3Param
	Color         Vec3Param
	renderType    int32
	Scale         float32
	DrawWireframe bool
	//Function 1
	Function        string
	errortext       string
	descriptionText string
	//lower & upper bound
	lower, upper float32
	//resolution
	spinePoints  int32 //Num of slices
	circlePoints int32 //Num of points that make up one slice

	numpoints int

	vertices [][3]float32
	//normals
	faces [][3]int

	vbo          uint32
	vao          uint32
	normalbuffer uint32
	//normals
}

func NewDiskGenerator(id int) *DiskRevolution {
	d := DiskRevolution{
		id: id,
		Translation: Vec3Param{
			Name: "Translation ## " + fmt.Sprint(id),
			min:  -10000,
			max:  10000,
			step: .01},
		Color: Vec3Param{
			Name: "Color ## " + fmt.Sprint(id),
			min:  0,
			max:  1,
			step: .01},
		renderType:      1,
		descriptionText: "Build to get volume",
		upper:           4,
		lower:           0,
		Scale:           1,
		spinePoints:     4,
		circlePoints:    20,
		Function:        ".5+.5*sin(x)",
	}
	return &d
}

func (dr *DiskRevolution) Build() g.Widget {
	return g.TreeNode(fmt.Sprintf("Object Function %d ## %d", dr.id, dr.id)).Layout(
		g.InputInt(fmt.Sprintf("Render Mode ## %d", dr.id), &dr.renderType),
		g.InputText(fmt.Sprintf("Function ## %d", dr.id), &dr.Function),
		g.Checkbox("Draw Wireframe", &dr.DrawWireframe),
		g.InputFloat("Lower", &dr.lower),
		g.InputFloat("Upper", &dr.upper),
		g.InputInt(fmt.Sprintf("# of Spine Points ## %d", dr.id), &dr.spinePoints),
		g.InputInt(fmt.Sprintf("# of Circle Points ## %d", dr.id), &dr.circlePoints),
		&dr.Translation,
		g.InputFloat(fmt.Sprintf("Scale ## %d", dr.id), &dr.Scale),
		&dr.Color,
		g.Button(fmt.Sprintf("Rebuild Object ## %d", dr.id)).OnClick(dr.CreateMesh),

		g.InputTextMultiline(fmt.Sprint("## info", dr.id), &dr.descriptionText).Flags(g.InputTextFlags_ReadOnly),
		g.Label(fmt.Sprint("Errors for ", dr.id, "\n", dr.errortext)).Color(&color.RGBA{
			R: 255,
			G: 10,
			B: 10,
			A: 255,
		}),
	)
}

func (dr *DiskRevolution) CreateMesh() {
	if dr.spinePoints < 3 {
		dr.spinePoints = 3
	}
	if dr.circlePoints < 3 {
		dr.circlePoints = 3
	}
	exefunc, err := parser.ParseExpression(dr.Function)
	if err != nil {
		dr.errortext = "Failed: " + err.Error()
	}
	//Get volume
	volumeExpression, _ := parser.ParseExpression("3.14159*(" + exefunc.String() + ")^2")
	//volume:=parser.
	volume := parser.Integrate(volumeExpression, map[string]float64{}, "x", float64(dr.lower), float64(dr.upper))
	dr.descriptionText = fmt.Sprintf("Volume was %g\n", volume)
	dr.vertices = make([][3]float32, 0, 3*dr.spinePoints*dr.circlePoints)
	dr.faces = [][3]int{}
	heights := make([]float32, dr.spinePoints)
	vars := map[string]float64{"x": 0}
	dt := (dr.upper - dr.lower) / float32(dr.spinePoints-3)
	xTotal := dr.lower
	//Create vertices
	for i := 0; i < len(heights); i++ {
		//0 and len()-1 are interior faces /edge loops

		//Create height
		x := xTotal
		vars["x"] = float64(x)
		heights[i] = float32(exefunc.Evaluate(vars))
		if i == 0 || i == len(heights)-1 {
			heights[i] *= .999
		}
		//Create ring
		dtheta := (2 * math.Pi) / float64(dr.circlePoints)
		for j := 0; j < int(dr.circlePoints); j++ {
			angle := dtheta * float64(j)
			y := math.Sin(angle) * float64(heights[i])
			z := math.Cos(angle) * float64(heights[i])
			dr.vertices = append(dr.vertices, [3]float32{float32(x), float32(y), float32(z)})
		}
		if i != 0 && i != len(heights)-2 {
			xTotal += dt
		}
	}
	//Create faces between rings

	//Link rings
	cp := int(dr.circlePoints)
	for segment := 0; segment < int(dr.spinePoints)-1; segment++ {
		for j := 0; j < int(dr.circlePoints); j++ {
			a := (j + 1)
			a %= cp

			face := [3]int{(segment+1)*cp + j, segment*cp + j, segment*cp + a}
			dr.faces = append(dr.faces, face)

			face2 := [3]int{(segment+1)*cp + a, (segment+1)*cp + j, segment*cp + a}

			dr.faces = append(dr.faces, face2)
		}
	}

	dr.vertices = append(dr.vertices, [3]float32{dr.lower, 0, 0})
	//last ring offset
	lro := int((dr.spinePoints - 1) * dr.circlePoints)

	frc := len(dr.vertices) - 1 //first ring center
	dr.vertices = append(dr.vertices, [3]float32{dr.upper, 0, 0})
	lrc := len(dr.vertices) - 1 //last ring center

	//Assemble rings into faces
	//Assemble base plate
	for i := 0; i < int(dr.circlePoints); i++ {
		face := [3]int{frc, i, (i + 1) % int(dr.circlePoints)}
		dr.faces = append(dr.faces, face)
	}
	//Assemble top plate
	for i := lro; i < lro+int(dr.circlePoints); i++ {
		final := i + 1
		if final == lro+int(dr.circlePoints) {
			final -= int(dr.circlePoints)
		}
		face := [3]int{lrc, i, final}
		dr.faces = append(dr.faces, face)
	}

	//Convert to points and normals
	var points []float32
	println("Num faces: ", len(dr.faces))
	println("Num Pouints: ", len(dr.vertices))

	//Normals, point away from center axis in direction of point
	norms := make([][3]float32, len(dr.vertices))
	normals := make([]float32, len(points))

	for _, face := range dr.faces {
		n := CalcNormal(face, dr.vertices)
		norms[face[0]] = Add(norms[face[0]], n)
		// Vertice 1
		v1 := dr.vertices[face[0]]
		points = append(points, v1[0], v1[1], v1[2])

		// Vertice 2
		v2 := dr.vertices[face[1]]
		points = append(points, v2[0], v2[1], v2[2])

		// Vertice 3
		v3 := dr.vertices[face[2]]
		points = append(points, v3[0], v3[1], v3[2])

	}
	for i := range norms {
		norms[i] = mgl32.Vec3(norms[i]).Normalize()
	}
	for _, face := range dr.faces {
		n1 := norms[face[0]]
		normals = append(normals, n1[0], n1[1], n1[2])
		n2 := norms[face[1]]
		normals = append(normals, n2[0], n2[1], n2[2])
		n3 := norms[face[2]]
		normals = append(normals, n3[0], n3[1], n3[2])
	}

	fmt.Println("points", points)
	//Create vbo, vao, normals

	{
		dr.numpoints = len(points)

		gl.GenBuffers(1, &dr.vbo)
		gl.BindBuffer(gl.ARRAY_BUFFER, dr.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

		gl.GenVertexArrays(1, &dr.vao)
		gl.BindVertexArray(dr.vao)
		gl.BindBuffer(gl.ARRAY_BUFFER, dr.vbo)

		gl.GenBuffers(1, &dr.normalbuffer)
		gl.BindBuffer(gl.ARRAY_BUFFER, dr.normalbuffer)
		gl.BufferData(gl.ARRAY_BUFFER, 4*len(normals), gl.Ptr(normals), gl.STATIC_DRAW)

		gl.EnableVertexAttribArray(0)
		gl.BindBuffer(gl.ARRAY_BUFFER, dr.vbo)
		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

		gl.EnableVertexAttribArray(1)
		gl.BindBuffer(gl.ARRAY_BUFFER, dr.normalbuffer)
		gl.VertexAttribPointer(1, 3, gl.FLOAT, true, 0, nil)
	}

}
func Add(a, b [3]float32) [3]float32 {
	return [3]float32{
		a[0] + b[0],
		a[1] + b[1],
		a[2] + b[2],
	}
}
func CalcNormal(face [3]int, verts [][3]float32) [3]float32 {
	A := mgl32.Vec3(verts[face[0]])
	B := mgl32.Vec3(verts[face[1]])
	C := mgl32.Vec3(verts[face[2]])

	U := B.Sub(A)
	V := C.Sub(A)
	return [3]float32{
		U.Y()*V.Z() - U.Z()*V.Y(),
		U.Z()*V.X() - U.X()*V.Z(),
		U.X()*V.Y() - U.Y()*V.X(),
	}

}
func (dr *DiskRevolution) Draw(projection, view mgl32.Mat4, program uint32) {
	modelMatrixName := "modelMatrix"
	viewMatrixName := "viewMatrix"
	projMatrixName := "projMatrix"
	mvpMatrixName := "MVP"

	gl.Disable(gl.CULL_FACE)

	// Set up model martix for shader
	model := mgl32.Ident4()

	model = mgl32.Translate3D(dr.Translation.Value[0], dr.Translation.Value[1], dr.Translation.Value[2])

	model = model.Mul4(mgl32.Scale3D(dr.Scale, dr.Scale, dr.Scale))

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
	gl.Uniform1i(ShadeNormPos, dr.renderType)

	MatColorPos := gl.GetUniformLocation(program, gl.Str("MaterialColor"+"\x00"))
	gl.Uniform3f(MatColorPos, dr.Color.Value[0], dr.Color.Value[1], dr.Color.Value[2])

	gl.Disable(gl.CULL_FACE)

	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
	gl.BindVertexArray(dr.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(dr.numpoints/3))

	//Draw Borders
	if dr.DrawWireframe {
		gl.Uniform1i(ShadeNormPos, 0)
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(dr.numpoints/3))
	}

	gl.Enable(gl.CULL_FACE)

}

//Find if they cross
//if they do; error
//find inner outer function
//f((low+upper)/2) < g((low+upper)/2)
