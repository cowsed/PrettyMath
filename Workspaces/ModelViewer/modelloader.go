package modelviewer

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
)

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
