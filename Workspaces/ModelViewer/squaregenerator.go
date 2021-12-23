package modelviewer

import (
	"fmt"
	"image/color"
	"math"

	g "github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"

	parser "github.com/cowsed/Parser"
)

type SquareExtrusion struct {
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
	spinePoints int32 //Num of slices

	numpoints int

	vertices [][3]float32
	//normals
	faces [][3]int

	vbo          uint32
	vao          uint32
	normalbuffer uint32
	//normals
}

func NewSquareGenerator(id int) *SquareExtrusion {
	d := SquareExtrusion{
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
		Function:        ".5+.5*sin(x)",
	}
	return &d
}
func (dr *SquareExtrusion) Dispose() {
	fmt.Println("Deleting buffers")
	gl.DeleteBuffers(1, &dr.normalbuffer)
	gl.DeleteBuffers(1, &dr.vbo)
	gl.DeleteVertexArrays(1, &dr.vao)
	dr.vertices = [][3]float32{}
}
func (dr *SquareExtrusion) Build() g.Widget {
	return g.TreeNode(fmt.Sprintf("Square Extrusion Function %d ## %d", dr.id, dr.id)).Layout(
		g.InputInt(fmt.Sprintf("Render Mode ## %d", dr.id), &dr.renderType),
		g.InputText(fmt.Sprintf("Function ## %d", dr.id), &dr.Function),
		g.Checkbox("Draw Wireframe", &dr.DrawWireframe),
		g.InputFloat("Lower", &dr.lower),
		g.InputFloat("Upper", &dr.upper),
		g.InputInt(fmt.Sprintf("# of Spine Points ## %d", dr.id), &dr.spinePoints),
		&dr.Translation,
		g.SliderFloat(fmt.Sprintf("Scale ## %d", dr.id), &dr.Scale, 0, 10),
		g.Custom(func() {
			rep := [3]float32{dr.Color.Value[0], dr.Color.Value[1], dr.Color.Value[2]}
			imgui.ColorEdit3(fmt.Sprintf("Color ## %d", dr.id), &rep)
			dr.Color.Value[0] = rep[0]
			dr.Color.Value[1] = rep[1]
			dr.Color.Value[2] = rep[2]
		}),
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

func (dr *SquareExtrusion) CalcVolume(exefunc parser.Expression) float64 {

	volumeExpression, _ := parser.ParseExpression("(" + exefunc.String() + ")^2")
	volume := parser.Integrate(volumeExpression, map[string]float64{}, "x", float64(dr.lower), float64(dr.upper))
	return volume
}
func (dr *SquareExtrusion) CalcSA(exefunc parser.Expression) (float64, float64, float64, float64) {
	//Need to do some calc to figure this out
	/*
		saLower := 2 * math.Pi * exefunc.Evaluate(map[string]float64{"x": float64(dr.lower)})
		if math.IsNaN(saLower) {
			fmt.Println("ERROR ON SALOWER")
			return math.NaN(), 0, 0, 0

		}
		saUpper := 2 * math.Pi * exefunc.Evaluate(map[string]float64{"x": float64(dr.upper)})
		if math.IsNaN(saUpper) {
			fmt.Println("ERROR ON SAHIGH")
			return math.NaN(), 0, 0, 0

		}
		dydx := exefunc.Derive("x").String()
		saExpression, err := parser.ParseExpression("2*3.14159265359*(" + exefunc.String() + ") * (1+(" + dydx + ")^2)^.5")
		if err != nil {
			return math.NaN(), 0, 0, 0
		}
		saSide := parser.Integrate(saExpression, map[string]float64{}, "x", float64(dr.lower), float64(dr.upper))
		if math.IsNaN(saSide) {
			fmt.Println("ERROR ON SIDE - which parsed to ", saExpression.String())
			return math.NaN(), 0, 0, 0

		}
		surfaceArea := saLower + saUpper + saSide
	*/
	//return saLower, saUpper, saSide, surfaceArea
	return math.NaN(), math.NaN(), math.NaN(), math.NaN()
}

func (se *SquareExtrusion) CreateMesh() {
	if se.spinePoints < 3 {
		se.spinePoints = 3
	}

	exefunc, err := parser.ParseExpression(se.Function)
	if err != nil {
		se.errortext = "Failed: " + err.Error()
		return
	}
	//Get volume
	//Volume was 50.26543999999994
	volume := se.CalcVolume(exefunc)
	saL, saU, saLat, surfaceArea := se.CalcSA(exefunc)
	se.descriptionText = fmt.Sprintf("Volume was %g\nSurface Area was %g \nLateral SA: %g\nUpper Surface Area: %g\nLower Surface Area: %g\n", volume, surfaceArea, saLat, saL, saU)
	se.vertices = make([][3]float32, 0, 3*se.spinePoints*4)
	se.faces = [][3]int{}
	heights := make([]float32, se.spinePoints)
	vars := map[string]float64{"x": 0}
	dt := (se.upper - se.lower) / float32(se.spinePoints-3)
	xTotal := se.lower
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
		//Create square
		s := heights[i]
		se.vertices = append(se.vertices, [3]float32{float32(x), float32(-s / 2), float32(-s / 2)})
		se.vertices = append(se.vertices, [3]float32{float32(x), float32(-s / 2), float32(s / 2)})
		se.vertices = append(se.vertices, [3]float32{float32(x), float32(s / 2), float32(s / 2)})
		se.vertices = append(se.vertices, [3]float32{float32(x), float32(s / 2), float32(-s / 2)})

		if i != 0 && i != len(heights)-2 {
			xTotal += dt
		}
	}
	//Create faces between rings

	//Link rings
	cp := int(4)
	for segment := 0; segment < int(se.spinePoints)-1; segment++ {
		for j := 0; j < int(4); j++ {
			a := (j + 1)
			a %= cp

			face := [3]int{(segment+1)*cp + j, segment*cp + j, segment*cp + a}
			se.faces = append(se.faces, face)

			face2 := [3]int{(segment+1)*cp + a, (segment+1)*cp + j, segment*cp + a}

			se.faces = append(se.faces, face2)
		}
	}
	//Assemble first face
	facea1 := [3]int{0, 1, 2}
	facea2 := [3]int{0, 2, 3}
	se.faces = append(se.faces, facea1, facea2)
	//Assemble last face
	lfo := len(se.vertices) - 4 //Last face offset
	faceb1 := [3]int{lfo + 0, lfo + 1, lfo + 2}
	faceb2 := [3]int{lfo + 0, lfo + 2, lfo + 3}
	se.faces = append(se.faces, faceb1, faceb2)

	//Convert to list of gl style points and normals
	//var points []float32
	println("Num faces: ", len(se.faces))
	println("Num Pouints: ", len(se.vertices))

	faceNormals := make([][3]float32, len(se.faces)) //duplicate points with correct normals

	pointsToDraw := make([][3]float32, 3*len(se.faces))
	NormalsToDraw := make([][3]float32, 3*len(se.faces))

	for i, face := range se.faces {
		n := CalcNormal(face, se.vertices)
		faceNormals[i] = n

		pointsToDraw[i*3] = se.vertices[se.faces[i][0]]
		pointsToDraw[i*3+1] = se.vertices[se.faces[i][1]]
		pointsToDraw[i*3+2] = se.vertices[se.faces[i][2]]
		NormalsToDraw[i*3] = faceNormals[i]
		NormalsToDraw[i*3+1] = faceNormals[i]
		NormalsToDraw[i*3+2] = faceNormals[i]
	}
	pointsRaw := make([]float32, 3*len(pointsToDraw)) //points stored as individual floats
	normsRaw := make([]float32, 3*len(NormalsToDraw)) //points stored as individual floats
	for i, p := range pointsToDraw {
		pointsRaw[i*3] = p[0]
		pointsRaw[i*3+1] = p[1]
		pointsRaw[i*3+2] = p[2]

		normsRaw[i*3] = NormalsToDraw[i][0]
		normsRaw[i*3+1] = NormalsToDraw[i][1]
		normsRaw[i*3+2] = NormalsToDraw[i][2]
	}

	//Delete old ones
	if se.vbo != 0 { //0 is default i think so dont delete it
		gl.DeleteBuffers(1, &se.vbo)
	}
	if se.normalbuffer != 0 { //0 is default i think so dont delete it
		gl.DeleteBuffers(1, &se.normalbuffer)
	}
	if se.vao != 0 { //0 is default i think so dont delete it
		gl.DeleteVertexArrays(1, &se.vao)
	}
	//Create vbo, vao, normals
	{
		se.numpoints = len(pointsRaw)

		gl.GenBuffers(1, &se.vbo)
		gl.BindBuffer(gl.ARRAY_BUFFER, se.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, 4*len(pointsRaw), gl.Ptr(pointsRaw), gl.STATIC_DRAW)

		gl.GenVertexArrays(1, &se.vao)
		gl.BindVertexArray(se.vao)
		gl.BindBuffer(gl.ARRAY_BUFFER, se.vbo)

		gl.GenBuffers(1, &se.normalbuffer)
		gl.BindBuffer(gl.ARRAY_BUFFER, se.normalbuffer)
		gl.BufferData(gl.ARRAY_BUFFER, 4*len(normsRaw), gl.Ptr(normsRaw), gl.STATIC_DRAW)

		gl.EnableVertexAttribArray(0)
		gl.BindBuffer(gl.ARRAY_BUFFER, se.vbo)
		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

		gl.EnableVertexAttribArray(1)
		gl.BindBuffer(gl.ARRAY_BUFFER, se.normalbuffer)
		gl.VertexAttribPointer(1, 3, gl.FLOAT, true, 0, nil)
	}

}

func (dr *SquareExtrusion) Draw(projection, view mgl32.Mat4, program uint32) {
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
func (se *SquareExtrusion) SetID(id int) {
	se.id = id
}
