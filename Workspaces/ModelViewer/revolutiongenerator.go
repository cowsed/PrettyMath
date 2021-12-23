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
		spinePoints:     20,
		circlePoints:    20,
		Function:        ".5+.5*sin(x)",
	}
	return &d
}
func (dr *DiskRevolution) SetID(id int) {
	dr.id = id
}
func (dr *DiskRevolution) Dispose() {
	fmt.Println("Deleting buffers")
	gl.DeleteBuffers(1, &dr.normalbuffer)
	gl.DeleteBuffers(1, &dr.vbo)
	gl.DeleteVertexArrays(1, &dr.vao)
	dr.vertices = [][3]float32{}
}
func (dr *DiskRevolution) Build() g.Widget {
	return g.TreeNode(fmt.Sprintf("Revolution Function %d ## %d", dr.id, dr.id)).Layout(
		g.InputInt(fmt.Sprintf("Render Mode ## %d", dr.id), &dr.renderType),
		g.InputText(fmt.Sprintf("Function ## %d", dr.id), &dr.Function),
		g.Checkbox("Draw Wireframe", &dr.DrawWireframe),
		g.InputFloat("Lower", &dr.lower),
		g.InputFloat("Upper", &dr.upper),
		g.InputInt(fmt.Sprintf("# of Spine Points ## %d", dr.id), &dr.spinePoints),
		g.InputInt(fmt.Sprintf("# of Circle Points ## %d", dr.id), &dr.circlePoints),
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

func (dr *DiskRevolution) CalcVolume(exefunc parser.Expression) float64 {

	volumeExpression, _ := parser.ParseExpression("3.14159265359*(" + exefunc.String() + ")^2")
	volume := parser.Integrate(volumeExpression, map[string]float64{}, "x", float64(dr.lower), float64(dr.upper))
	return volume
}
func (dr *DiskRevolution) CalcSA(exefunc parser.Expression) (float64, float64, float64, float64) {

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
	return saLower, saUpper, saSide, surfaceArea
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
		return
	}
	//Get volume
	//Volume was 50.26543999999994
	volume := dr.CalcVolume(exefunc)
	saL, saU, saLat, surfaceArea := dr.CalcSA(exefunc)
	dr.descriptionText = fmt.Sprintf("Volume was %g\nSurface Area was %g \nLateral SA: %g\nUpper Surface Area: %g\nLower Surface Area: %g\n", volume, surfaceArea, saLat, saL, saU)
	dr.vertices = make([][3]float32, 0, 3*dr.spinePoints*dr.circlePoints)
	dr.faces = [][3]int{}
	heights := make([]float32, dr.spinePoints)
	vars := map[string]float64{"x": 0}
	dt := (dr.upper - dr.lower) / float32(dr.spinePoints-1)
	xTotal := dr.lower
	//Create vertices
	for i := 0; i < len(heights); i++ {
		//0 and len()-1 are interior faces /edge loops

		//Create height
		x := xTotal
		vars["x"] = float64(x)
		heights[i] = float32(exefunc.Evaluate(vars))

		//Create ring
		dtheta := (2 * math.Pi) / float64(dr.circlePoints)
		for j := 0; j < int(dr.circlePoints); j++ {
			angle := dtheta * float64(j)
			y := math.Sin(angle) * float64(heights[i])
			z := math.Cos(angle) * float64(heights[i])
			dr.vertices = append(dr.vertices, [3]float32{float32(x), float32(y), float32(z)})
		}
		xTotal += dt

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

	basePlatePoints := [][3]float32{}
	basePlateNorms := [][3]float32{}

	//Lower Base Plate Rings
	dtheta := (2 * math.Pi) / float64(dr.circlePoints)
	for j := 0; j < int(dr.circlePoints); j++ {
		angle := dtheta * float64(j)
		y := math.Sin(angle) * float64(heights[0])
		z := math.Cos(angle) * float64(heights[0])
		basePlatePoints = append(basePlatePoints, [3]float32{float32(dr.lower), float32(y), float32(z)})
		basePlateNorms = append(basePlateNorms, [3]float32{-1, 0, 0})
	}
	//Upper Base Plate Rings
	lastHeight := len(heights) - 1
	for j := 0; j < int(dr.circlePoints); j++ {
		angle := dtheta * float64(j)
		y := math.Sin(angle) * float64(heights[lastHeight])
		z := math.Cos(angle) * float64(heights[lastHeight])
		basePlatePoints = append(basePlatePoints, [3]float32{float32(dr.upper), float32(y), float32(z)})
		basePlateNorms = append(basePlateNorms, [3]float32{1, 0, 0})
	}
	//Lower Base plate faces
	lowerBPStart := len(dr.vertices)
	for j := 1; j < int(dr.circlePoints)-1; j++ {
		face := [3]int{lowerBPStart, lowerBPStart + j, lowerBPStart + j + 1}
		dr.faces = append(dr.faces, face)
	}
	upperBPStart := len(dr.vertices) + int(dr.circlePoints)
	for j := 1; j < int(dr.circlePoints)-1; j++ {
		face := [3]int{upperBPStart, upperBPStart + j, upperBPStart + j + 1}
		dr.faces = append(dr.faces, face)
	}

	dr.vertices = append(dr.vertices, basePlatePoints...)

	//Convert to pointsRaw and normals
	var pointsRaw []float32
	println("Num faces: ", len(dr.faces))
	println("Num Pouints: ", len(dr.vertices))

	//Normals, point away from center axis in direction of point
	norms := make([][3]float32, len(dr.vertices))
	normalsRaw := make([]float32, len(pointsRaw))

	for _, face := range dr.faces {
		n := CalcNormal(face, dr.vertices)
		norms[face[0]] = Add(norms[face[0]], n)
		norms[face[1]] = Add(norms[face[1]], n)
		norms[face[2]] = Add(norms[face[2]], n)

		// Vertice 1
		v1 := dr.vertices[face[0]]
		pointsRaw = append(pointsRaw, v1[0], v1[1], v1[2])

		// Vertice 2
		v2 := dr.vertices[face[1]]
		pointsRaw = append(pointsRaw, v2[0], v2[1], v2[2])

		// Vertice 3
		v3 := dr.vertices[face[2]]
		pointsRaw = append(pointsRaw, v3[0], v3[1], v3[2])

	}
	norms = append(norms, basePlateNorms...)
	for i := range norms {

		norms[i] = mgl32.Vec3(norms[i]).Normalize()
	}
	for _, face := range dr.faces {
		n1 := norms[face[0]]
		normalsRaw = append(normalsRaw, n1[0], n1[1], n1[2])
		n2 := norms[face[1]]
		normalsRaw = append(normalsRaw, n2[0], n2[1], n2[2])
		n3 := norms[face[2]]
		normalsRaw = append(normalsRaw, n3[0], n3[1], n3[2])
	}

	//Delete old ones
	if dr.vbo != 0 { //0 is default i think so dont delete it
		gl.DeleteBuffers(1, &dr.vbo)
	}
	if dr.normalbuffer != 0 { //0 is default i think so dont delete it
		gl.DeleteBuffers(1, &dr.normalbuffer)
	}
	if dr.vao != 0 { //0 is default i think so dont delete it
		gl.DeleteVertexArrays(1, &dr.vao)
	}
	//Create vbo, vao, normals
	{
		dr.numpoints = len(pointsRaw)

		gl.GenBuffers(1, &dr.vbo)
		gl.BindBuffer(gl.ARRAY_BUFFER, dr.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, 4*len(pointsRaw), gl.Ptr(pointsRaw), gl.STATIC_DRAW)

		gl.GenVertexArrays(1, &dr.vao)
		gl.BindVertexArray(dr.vao)
		gl.BindBuffer(gl.ARRAY_BUFFER, dr.vbo)

		gl.GenBuffers(1, &dr.normalbuffer)
		gl.BindBuffer(gl.ARRAY_BUFFER, dr.normalbuffer)
		gl.BufferData(gl.ARRAY_BUFFER, 4*len(normalsRaw), gl.Ptr(normalsRaw), gl.STATIC_DRAW)

		gl.EnableVertexAttribArray(0)
		gl.BindBuffer(gl.ARRAY_BUFFER, dr.vbo)
		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

		gl.EnableVertexAttribArray(1)
		gl.BindBuffer(gl.ARRAY_BUFFER, dr.normalbuffer)
		gl.VertexAttribPointer(1, 3, gl.FLOAT, true, 0, nil)
	}

}

//Checks if a face is from a base plate
//want to do this so it does not take it into account when calculating side normals
func CheckNotInPlate(i1, i2, i3 int, circlePoints, totalPoints int) bool {
	if i1 < circlePoints && i2 < circlePoints && i3 < circlePoints { //is  in base plate
		fmt.Println("was in bas eplate. ", i1, i2, i3)
		return false //Is in the base plate
	}
	endCutoff := totalPoints - circlePoints - 2
	if i1 > endCutoff && i2 > endCutoff && i3 > endCutoff {
		fmt.Println("was in bas eplate. ", i1, i2, i3)
		return false //is in the top plate or in the center vertex
	}
	return true //True, is not from the base plate
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
	return mgl32.Vec3{
		U.Y()*V.Z() - U.Z()*V.Y(),
		U.Z()*V.X() - U.X()*V.Z(),
		U.X()*V.Y() - U.Y()*V.X(),
	}.Normalize()

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
